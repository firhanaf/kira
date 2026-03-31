package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/kira-app/kira-server/internal/domain"
	"github.com/kira-app/kira-server/internal/middleware"
	"github.com/kira-app/kira-server/internal/service"
)

type ProjectHandler struct {
	svc *service.ProjectService
}

func NewProjectHandler(svc *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{svc: svc}
}

func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		ClientName  string `json:"client_name"`
		ClientEmail string `json:"client_email"`
		Currency    string `json:"currency"`
		HourlyRate  int    `json:"hourly_rate"`
		Color       string `json:"color"`
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	p, err := h.svc.Create(r.Context(), userID, service.CreateProjectInput{
		Name:        req.Name,
		Description: req.Description,
		ClientName:  req.ClientName,
		ClientEmail: req.ClientEmail,
		Currency:    req.Currency,
		HourlyRate:  req.HourlyRate,
		Color:       req.Color,
	})
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respond(w, http.StatusCreated, p)
}

func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	projects, err := h.svc.List(r.Context(), userID)
	if err != nil {
		respondDomainError(w, err)
		return
	}
	if projects == nil {
		projects = []domain.Project{}
	}
	respond(w, http.StatusOK, projects)
}

func (h *ProjectHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	p, err := h.svc.Get(r.Context(), id, userID)
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respond(w, http.StatusOK, p)
}

func (h *ProjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		ClientName  string `json:"client_name"`
		ClientEmail string `json:"client_email"`
		Status      string `json:"status"`
		Currency    string `json:"currency"`
		HourlyRate  int    `json:"hourly_rate"`
		Color       string `json:"color"`
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	p, err := h.svc.Update(r.Context(), id, userID, service.UpdateProjectInput{
		Name:        req.Name,
		Description: req.Description,
		ClientName:  req.ClientName,
		ClientEmail: req.ClientEmail,
		Status:      req.Status,
		Currency:    req.Currency,
		HourlyRate:  req.HourlyRate,
		Color:       req.Color,
	})
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respond(w, http.StatusOK, p)
}

func (h *ProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	if err := h.svc.Delete(r.Context(), id, userID); err != nil {
		respondDomainError(w, err)
		return
	}
	respond(w, http.StatusOK, map[string]string{"message": "deleted"})
}
