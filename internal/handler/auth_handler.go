package handler

import (
	"net/http"

	"github.com/kira-app/kira-server/internal/domain"
	"github.com/kira-app/kira-server/internal/middleware"
	"github.com/kira-app/kira-server/internal/service"
)

type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Name     string `json:"name"`
		Password string `json:"password"`
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.svc.Register(r.Context(), service.RegisterInput{
		Email:    req.Email,
		Name:     req.Name,
		Password: req.Password,
	})
	if err != nil {
		respondDomainError(w, err)
		return
	}

	respond(w, http.StatusCreated, map[string]any{
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
		"user":          userResponse(result.User),
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.svc.Login(r.Context(), service.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		respondDomainError(w, err)
		return
	}

	respond(w, http.StatusOK, map[string]any{
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
		"user":          userResponse(result.User),
	})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := decodeJSON(r, &req); err != nil || req.RefreshToken == "" {
		respondError(w, http.StatusBadRequest, "refresh_token is required")
		return
	}

	result, err := h.svc.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		respondDomainError(w, err)
		return
	}

	respond(w, http.StatusOK, map[string]any{
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := decodeJSON(r, &req); err != nil || req.RefreshToken == "" {
		respondError(w, http.StatusBadRequest, "refresh_token is required")
		return
	}

	_ = h.svc.Logout(r.Context(), req.RefreshToken)
	respond(w, http.StatusOK, map[string]string{"message": "logged out"})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	u, err := h.svc.GetUser(r.Context(), userID)
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respond(w, http.StatusOK, userResponse(u))
}

func (h *AuthHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	u, err := h.svc.GetUser(r.Context(), userID)
	if err != nil {
		respondDomainError(w, err)
		return
	}

	var req struct {
		Name            string `json:"name"`
		AvatarURL       string `json:"avatar_url"`
		DefaultCurrency string `json:"default_currency"`
		DefaultRate     int    `json:"default_rate"`
		Timezone        string `json:"timezone"`
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name != "" {
		u.Name = req.Name
	}
	if req.AvatarURL != "" {
		u.AvatarURL = req.AvatarURL
	}
	if req.DefaultCurrency != "" {
		u.DefaultCurrency = req.DefaultCurrency
	}
	if req.DefaultRate > 0 {
		u.DefaultRate = req.DefaultRate
	}
	if req.Timezone != "" {
		u.Timezone = req.Timezone
	}

	if err := h.svc.UpdateUser(r.Context(), u); err != nil {
		respondDomainError(w, err)
		return
	}
	respond(w, http.StatusOK, userResponse(u))
}

func userResponse(u *domain.User) map[string]any {
	return map[string]any{
		"id":               u.ID,
		"email":            u.Email,
		"name":             u.Name,
		"avatar_url":       u.AvatarURL,
		"default_currency": u.DefaultCurrency,
		"default_rate":     u.DefaultRate,
		"timezone":         u.Timezone,
		"created_at":       u.CreatedAt,
	}
}
