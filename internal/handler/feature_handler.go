package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/kira-app/kira-server/internal/domain"
	"github.com/kira-app/kira-server/internal/middleware"
	"github.com/kira-app/kira-server/internal/service"
)

type FeatureHandler struct {
	svc *service.FeatureService
}

func NewFeatureHandler(svc *service.FeatureService) *FeatureHandler {
	return &FeatureHandler{svc: svc}
}

func (h *FeatureHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	projectID, err := uuid.Parse(r.PathValue("projectID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Notes       string `json:"notes"`
		Severity    string `json:"severity"`
		GitBranch   string `json:"git_branch"`
		GitRepoURL  string `json:"git_repo_url"`
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	f, err := h.svc.Create(r.Context(), projectID, userID, service.CreateFeatureInput{
		Name:        req.Name,
		Description: req.Description,
		Notes:       req.Notes,
		Severity:    req.Severity,
		GitBranch:   req.GitBranch,
		GitRepoURL:  req.GitRepoURL,
	})
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respond(w, http.StatusCreated, f)
}

func (h *FeatureHandler) ListByProject(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	projectID, err := uuid.Parse(r.PathValue("projectID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	features, err := h.svc.ListByProject(r.Context(), projectID, userID)
	if err != nil {
		respondDomainError(w, err)
		return
	}
	if features == nil {
		features = []domain.Feature{}
	}
	respond(w, http.StatusOK, features)
}

func (h *FeatureHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid feature id")
		return
	}

	f, err := h.svc.Get(r.Context(), id, userID)
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respond(w, http.StatusOK, f)
}

func (h *FeatureHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid feature id")
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Notes       string `json:"notes"`
		Severity    string `json:"severity"`
		Status      string `json:"status"`
		GitBranch   string `json:"git_branch"`
		GitRepoURL  string `json:"git_repo_url"`
		Position    int    `json:"position"`
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	f, err := h.svc.Update(r.Context(), id, userID, service.UpdateFeatureInput{
		Name:        req.Name,
		Description: req.Description,
		Notes:       req.Notes,
		Severity:    req.Severity,
		Status:      req.Status,
		GitBranch:   req.GitBranch,
		GitRepoURL:  req.GitRepoURL,
		Position:    req.Position,
	})
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respond(w, http.StatusOK, f)
}

func (h *FeatureHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid feature id")
		return
	}

	if err := h.svc.Delete(r.Context(), id, userID); err != nil {
		respondDomainError(w, err)
		return
	}
	respond(w, http.StatusOK, map[string]string{"message": "deleted"})
}
