package handler

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/kira-app/kira-server/internal/middleware"
	"github.com/kira-app/kira-server/internal/service"
)

type NotificationHandler struct {
	svc *service.NotificationService
}

func NewNotificationHandler(svc *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

func (h *NotificationHandler) SendEmail(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	var req struct {
		ProjectID      string  `json:"project_id"`
		FeatureID      *string `json:"feature_id"`
		RecipientEmail string  `json:"recipient_email"`
		Subject        string  `json:"subject"`
		Message        string  `json:"message"`
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	projectID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid project_id")
		return
	}

	var featureID *uuid.UUID
	if req.FeatureID != nil {
		fid, err := uuid.Parse(*req.FeatureID)
		if err == nil {
			featureID = &fid
		}
	}

	n, err := h.svc.SendEmail(r.Context(), userID, service.SendEmailInput{
		ProjectID:      projectID,
		FeatureID:      featureID,
		RecipientEmail: req.RecipientEmail,
		Subject:        req.Subject,
		Message:        req.Message,
	})
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respond(w, http.StatusCreated, n)
}

func (h *NotificationHandler) CreateShareableLink(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	var req struct {
		ProjectID      string  `json:"project_id"`
		FeatureID      *string `json:"feature_id"`
		Message        string  `json:"message"`
		ExpiresInHours int     `json:"expires_in_hours"`
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	projectID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid project_id")
		return
	}

	var featureID *uuid.UUID
	if req.FeatureID != nil {
		fid, err := uuid.Parse(*req.FeatureID)
		if err == nil {
			featureID = &fid
		}
	}

	if req.ExpiresInHours == 0 {
		req.ExpiresInHours = 72
	}

	n, err := h.svc.CreateShareableLink(r.Context(), userID, service.CreateShareableLinkInput{
		ProjectID: projectID,
		FeatureID: featureID,
		Message:   req.Message,
		ExpiresIn: time.Duration(req.ExpiresInHours) * time.Hour,
	})
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respond(w, http.StatusCreated, n)
}

func (h *NotificationHandler) GetByShareToken(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	if token == "" {
		respondError(w, http.StatusBadRequest, "token is required")
		return
	}

	n, err := h.svc.GetByShareToken(r.Context(), token)
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respond(w, http.StatusOK, n)
}
