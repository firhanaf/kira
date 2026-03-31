package handler

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/kira-app/kira-server/internal/domain"
	"github.com/kira-app/kira-server/internal/middleware"
	"github.com/kira-app/kira-server/internal/service"
)

type InvoiceHandler struct {
	svc    *service.InvoiceService
	appURL string
}

func NewInvoiceHandler(svc *service.InvoiceService, appURL string) *InvoiceHandler {
	return &InvoiceHandler{svc: svc, appURL: appURL}
}

func (h *InvoiceHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	projectID, err := uuid.Parse(r.PathValue("projectID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	var req struct {
		FeatureIDs []string   `json:"feature_ids"`
		MarginPct  float64    `json:"margin_pct"`
		TaxPct     float64    `json:"tax_pct"`
		DueDate    *time.Time `json:"due_date"`
		Notes      string     `json:"notes"`
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var featureIDs []uuid.UUID
	for _, idStr := range req.FeatureIDs {
		fid, err := uuid.Parse(idStr)
		if err != nil {
			continue
		}
		featureIDs = append(featureIDs, fid)
	}

	inv, err := h.svc.Create(r.Context(), userID, service.CreateInvoiceInput{
		ProjectID:  projectID,
		FeatureIDs: featureIDs,
		MarginPct:  req.MarginPct,
		TaxPct:     req.TaxPct,
		DueDate:    req.DueDate,
		Notes:      req.Notes,
	})
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respond(w, http.StatusCreated, inv)
}

func (h *InvoiceHandler) ListByProject(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	projectID, err := uuid.Parse(r.PathValue("projectID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	invoices, err := h.svc.ListByProject(r.Context(), projectID, userID)
	if err != nil {
		respondDomainError(w, err)
		return
	}
	if invoices == nil {
		invoices = []domain.Invoice{}
	}
	respond(w, http.StatusOK, invoices)
}

func (h *InvoiceHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid invoice id")
		return
	}

	inv, err := h.svc.Get(r.Context(), id, userID)
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respond(w, http.StatusOK, inv)
}

func (h *InvoiceHandler) MarkSent(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid invoice id")
		return
	}

	inv, err := h.svc.MarkSent(r.Context(), id, userID)
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respond(w, http.StatusOK, inv)
}

func (h *InvoiceHandler) MarkPaid(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid invoice id")
		return
	}

	inv, err := h.svc.MarkPaid(r.Context(), id, userID)
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respond(w, http.StatusOK, inv)
}

func (h *InvoiceHandler) CreateShareLink(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid invoice id")
		return
	}

	var req struct {
		ExpiresInHours int `json:"expires_in_hours"`
	}
	_ = decodeJSON(r, &req)
	if req.ExpiresInHours == 0 {
		req.ExpiresInHours = 72 // 3 days default
	}

	inv, err := h.svc.CreateShareLink(r.Context(), id, userID, time.Duration(req.ExpiresInHours)*time.Hour)
	if err != nil {
		respondDomainError(w, err)
		return
	}

	respond(w, http.StatusOK, map[string]any{
		"share_token":      inv.ShareToken,
		"share_expires_at": inv.ShareExpiresAt,
		"share_url":        h.appURL + "/public/invoices/" + inv.ShareToken,
	})
}

func (h *InvoiceHandler) GetPublic(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	if token == "" {
		respondError(w, http.StatusBadRequest, "token is required")
		return
	}

	inv, err := h.svc.GetByShareToken(r.Context(), token)
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respond(w, http.StatusOK, inv)
}

func (h *InvoiceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid invoice id")
		return
	}

	if err := h.svc.Delete(r.Context(), id, userID); err != nil {
		respondDomainError(w, err)
		return
	}
	respond(w, http.StatusOK, map[string]string{"message": "deleted"})
}
