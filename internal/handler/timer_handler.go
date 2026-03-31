package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/kira-app/kira-server/internal/middleware"
	"github.com/kira-app/kira-server/internal/service"
)

type TimerHandler struct {
	svc *service.TimerService
}

func NewTimerHandler(svc *service.TimerService) *TimerHandler {
	return &TimerHandler{svc: svc}
}

func (h *TimerHandler) Start(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	featureID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid feature id")
		return
	}

	var req struct {
		Note string `json:"note"`
	}
	_ = decodeJSON(r, &req)

	entry, err := h.svc.Start(r.Context(), featureID, userID, req.Note)
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respond(w, http.StatusCreated, entry)
}

func (h *TimerHandler) Stop(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	entry, err := h.svc.Stop(r.Context(), userID)
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respond(w, http.StatusOK, entry)
}

func (h *TimerHandler) Active(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	entry, err := h.svc.GetActive(r.Context(), userID)
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respond(w, http.StatusOK, entry)
}

func (h *TimerHandler) ListByFeature(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	featureID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid feature id")
		return
	}

	entries, err := h.svc.ListByFeature(r.Context(), featureID, userID)
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respond(w, http.StatusOK, entries)
}
