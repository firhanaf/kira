package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kira-app/kira-server/internal/config"
	"github.com/kira-app/kira-server/internal/db"
	"github.com/kira-app/kira-server/internal/email"
	"github.com/kira-app/kira-server/internal/handler"
	"github.com/kira-app/kira-server/internal/middleware"
	"github.com/kira-app/kira-server/internal/pdf"
	"github.com/kira-app/kira-server/internal/repository"
	"github.com/kira-app/kira-server/internal/service"
	"github.com/kira-app/kira-server/internal/sse"
	"github.com/kira-app/kira-server/pkg/token"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "fatal:", err)
		os.Exit(1)
	}
}

func run() error {
	// CONFIG
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// DATABASE
	ctx := context.Background()
	pool, err := db.New(ctx, cfg.DB)
	if err != nil {
		return fmt.Errorf("connect db: %w", err)
	}
	defer pool.Close()

	// RUN MIGRATIONS AUTOMATICALLY
	if err := db.Migrate(cfg.DB.URL, "./internal/db/migrations"); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	fmt.Println("Database migration applied")

	// INFRASTRUCTURE
	tokenMgr := token.NewManager(cfg.JWT.Secret, cfg.JWT.AccessExpiry, cfg.JWT.RefreshExpiry)
	sseHub := sse.New()
	emailSender := email.NewSender(cfg.SMTP)
	pdfGen := pdf.NewGenerator(cfg.Storage, cfg.App.URL)

	// REPOSITORIES
	userRepo := repository.NewUserRepo(pool)
	tokenRepo := repository.NewRefreshTokenRepo(pool)
	projectRepo := repository.NewProjectRepo(pool)
	featureRepo := repository.NewFeatureRepo(pool)
	timerRepo := repository.NewTimerRepo(pool)
	invoiceRepo := repository.NewInvoiceRepo(pool)
	notifRepo := repository.NewNotificationRepo(pool)

	// SERVICES
	authSvc := service.NewAuthService(userRepo, tokenRepo, tokenMgr)
	projectSvc := service.NewProjectService(projectRepo)
	featureSvc := service.NewFeatureService(featureRepo, projectRepo)
	timerSvc := service.NewTimerService(timerRepo, featureRepo, projectRepo, sseHub)
	invoiceSvc := service.NewInvoiceService(invoiceRepo, projectRepo, featureRepo, userRepo, pdfGen)
	notifSvc := service.NewNotificationService(notifRepo, projectRepo, emailSender, cfg.App.URL)

	// HANDLERS
	authH := handler.NewAuthHandler(authSvc)
	projectH := handler.NewProjectHandler(projectSvc)
	featureH := handler.NewFeatureHandler(featureSvc)
	timerH := handler.NewTimerHandler(timerSvc)
	invoiceH := handler.NewInvoiceHandler(invoiceSvc, cfg.App.URL)
	sseH := handler.NewSSEHandler(sseHub)
	notifH := handler.NewNotificationHandler(notifSvc)

	// ROUTER
	mux := http.NewServeMux()

	// HEALTH CHECK
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok", "env": cfg.App.Env})
	})

	// PUBLIC ROUTES
	mux.HandleFunc("POST /api/v1/auth/register", authH.Register)
	mux.HandleFunc("POST /api/v1/auth/login", authH.Login)
	mux.HandleFunc("POST /api/v1/auth/refresh", authH.Refresh)
	mux.HandleFunc("POST /api/v1/auth/logout", authH.Logout)
	mux.HandleFunc("GET /public/invoices/{token}", invoiceH.GetPublic)
	mux.HandleFunc("GET /public/notifications/{token}", notifH.GetByShareToken)

	// STATIC FILES (invoice PDFs/HTML)
	mux.Handle("GET /storage/", http.StripPrefix("/storage/", http.FileServer(http.Dir(cfg.Storage.LocalPath))))

	// PROTECTED ROUTES — wrap individually with auth middleware
	authMw := middleware.Auth(tokenMgr)

	// Me
	mux.Handle("GET /api/v1/me", authMw(http.HandlerFunc(authH.Me)))
	mux.Handle("PUT /api/v1/me", authMw(http.HandlerFunc(authH.UpdateMe)))

	// SSE
	mux.Handle("GET /api/v1/events", authMw(http.HandlerFunc(sseH.Stream)))

	// Projects
	mux.Handle("GET /api/v1/projects", authMw(http.HandlerFunc(projectH.List)))
	mux.Handle("POST /api/v1/projects", authMw(http.HandlerFunc(projectH.Create)))
	mux.Handle("GET /api/v1/projects/{id}", authMw(http.HandlerFunc(projectH.Get)))
	mux.Handle("PUT /api/v1/projects/{id}", authMw(http.HandlerFunc(projectH.Update)))
	mux.Handle("DELETE /api/v1/projects/{id}", authMw(http.HandlerFunc(projectH.Delete)))

	// Features
	mux.Handle("GET /api/v1/projects/{projectID}/features", authMw(http.HandlerFunc(featureH.ListByProject)))
	mux.Handle("POST /api/v1/projects/{projectID}/features", authMw(http.HandlerFunc(featureH.Create)))
	mux.Handle("GET /api/v1/features/{id}", authMw(http.HandlerFunc(featureH.Get)))
	mux.Handle("PUT /api/v1/features/{id}", authMw(http.HandlerFunc(featureH.Update)))
	mux.Handle("DELETE /api/v1/features/{id}", authMw(http.HandlerFunc(featureH.Delete)))

	// Timers
	mux.Handle("POST /api/v1/features/{id}/timer/start", authMw(http.HandlerFunc(timerH.Start)))
	mux.Handle("POST /api/v1/timer/stop", authMw(http.HandlerFunc(timerH.Stop)))
	mux.Handle("GET /api/v1/timer/active", authMw(http.HandlerFunc(timerH.Active)))
	mux.Handle("GET /api/v1/features/{id}/entries", authMw(http.HandlerFunc(timerH.ListByFeature)))

	// Invoices
	mux.Handle("GET /api/v1/projects/{projectID}/invoices", authMw(http.HandlerFunc(invoiceH.ListByProject)))
	mux.Handle("POST /api/v1/projects/{projectID}/invoices", authMw(http.HandlerFunc(invoiceH.Create)))
	mux.Handle("GET /api/v1/invoices/{id}", authMw(http.HandlerFunc(invoiceH.Get)))
	mux.Handle("DELETE /api/v1/invoices/{id}", authMw(http.HandlerFunc(invoiceH.Delete)))
	mux.Handle("POST /api/v1/invoices/{id}/send", authMw(http.HandlerFunc(invoiceH.MarkSent)))
	mux.Handle("POST /api/v1/invoices/{id}/paid", authMw(http.HandlerFunc(invoiceH.MarkPaid)))
	mux.Handle("POST /api/v1/invoices/{id}/share", authMw(http.HandlerFunc(invoiceH.CreateShareLink)))

	// Notifications
	mux.Handle("POST /api/v1/notifications/email", authMw(http.HandlerFunc(notifH.SendEmail)))
	mux.Handle("POST /api/v1/notifications/share", authMw(http.HandlerFunc(notifH.CreateShareableLink)))

	// GLOBAL MIDDLEWARE (apply outermost last)
	var h http.Handler = mux
	h = middleware.CORS(cfg.App.FrontendURL)(h)
	h = middleware.Logger()(h)

	// SERVER
	srv := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      h,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// GRACEFUL SHUTDOWN
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		fmt.Printf("Kira server running on http://localhost:%s (env: %s)\n", cfg.App.Port, cfg.App.Env)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintln(os.Stderr, "server error:", err)
		}
	}()
	<-quit
	fmt.Println("\nShutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}
