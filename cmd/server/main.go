package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kira-app/kira-server/internal/config"
	"github.com/kira-app/kira-server/internal/db"
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
	emailSender := newEmailSender(cfg.SMTP)
	pdfGenerator := newPDFGenerator(cfg.Storage)

	// REPOSITORIES

	// SERVICES

	// HANDLERS

	// ROUTER

	// GLOBAL MIDDLEWARE

	// HEALTH CHECK

	// SERVER
	srv := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      r,
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
