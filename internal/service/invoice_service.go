package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kira-app/kira-server/internal/domain"
	"github.com/kira-app/kira-server/internal/pdf"
	"github.com/kira-app/kira-server/internal/repository"
	"github.com/kira-app/kira-server/pkg/currency"
)

type InvoiceService struct {
	invoiceRepo *repository.InvoiceRepo
	projectRepo *repository.ProjectRepo
	featureRepo *repository.FeatureRepo
	userRepo    *repository.UserRepo
	pdfGen      *pdf.Generator
}

func NewInvoiceService(
	invoiceRepo *repository.InvoiceRepo,
	projectRepo *repository.ProjectRepo,
	featureRepo *repository.FeatureRepo,
	userRepo *repository.UserRepo,
	pdfGen *pdf.Generator,
) *InvoiceService {
	return &InvoiceService{
		invoiceRepo: invoiceRepo,
		projectRepo: projectRepo,
		featureRepo: featureRepo,
		userRepo:    userRepo,
		pdfGen:      pdfGen,
	}
}

type CreateInvoiceInput struct {
	ProjectID  uuid.UUID
	FeatureIDs []uuid.UUID
	MarginPct  float64
	TaxPct     float64
	DueDate    *time.Time
	Notes      string
}

func (s *InvoiceService) Create(ctx context.Context, userID uuid.UUID, in CreateInvoiceInput) (*domain.Invoice, error) {
	proj, err := s.projectRepo.GetByID(ctx, in.ProjectID, userID)
	if err != nil {
		return nil, err
	}

	if len(in.FeatureIDs) == 0 {
		return nil, fmt.Errorf("%w: at least one feature required", domain.ErrBadRequest)
	}

	var lineItems []domain.InvoiceLineItem
	subtotal := 0

	for _, fid := range in.FeatureIDs {
		f, err := s.featureRepo.GetByID(ctx, fid)
		if err != nil {
			continue
		}
		rate := proj.HourlyRate
		if rate == 0 {
			u, err := s.userRepo.GetByID(ctx, userID)
			if err == nil {
				rate = u.DefaultRate
			}
		}
		amount := currency.CalcAmount(f.ElapsedSeconds, rate)
		fid := fid
		li := domain.InvoiceLineItem{
			ID:             uuid.New(),
			FeatureID:      &fid,
			FeatureName:    f.Name,
			Severity:       f.Severity,
			GitBranch:      f.GitBranch,
			ElapsedSeconds: f.ElapsedSeconds,
			HourlyRate:     rate,
			Amount:         amount,
		}
		lineItems = append(lineItems, li)
		subtotal += amount
	}

	if in.TaxPct == 0 {
		in.TaxPct = 11 // default VAT
	}

	_, _, total := currency.CalcTotal(subtotal, in.MarginPct, in.TaxPct)

	inv := &domain.Invoice{
		ID:            uuid.New(),
		ProjectID:     in.ProjectID,
		UserID:        userID,
		InvoiceNumber: generateInvoiceNumber(),
		Status:        domain.InvoiceStatusDraft,
		Subtotal:      subtotal,
		MarginPct:     in.MarginPct,
		TaxPct:        in.TaxPct,
		Total:         total,
		Currency:      proj.Currency,
		DueDate:       in.DueDate,
		Notes:         in.Notes,
		LineItems:     lineItems,
	}

	if err := s.invoiceRepo.Create(ctx, inv); err != nil {
		return nil, err
	}

	// Generate invoice document in background (non-fatal if it fails)
	u, err := s.userRepo.GetByID(ctx, userID)
	if err == nil {
		if pdfURL, err := s.pdfGen.GenerateInvoice(inv, proj, u); err == nil {
			inv.PDFUrl = pdfURL
			_ = s.invoiceRepo.Update(ctx, inv)
		}
	}

	return inv, nil
}

func (s *InvoiceService) Get(ctx context.Context, id, userID uuid.UUID) (*domain.Invoice, error) {
	inv, err := s.invoiceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if inv.UserID != userID {
		return nil, domain.ErrForbidden
	}
	return inv, nil
}

func (s *InvoiceService) GetByShareToken(ctx context.Context, token string) (*domain.Invoice, error) {
	inv, err := s.invoiceRepo.GetByShareToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if inv.ShareExpiresAt != nil && time.Now().After(*inv.ShareExpiresAt) {
		return nil, fmt.Errorf("%w: share link expired", domain.ErrNotFound)
	}
	return inv, nil
}

func (s *InvoiceService) ListByProject(ctx context.Context, projectID, userID uuid.UUID) ([]domain.Invoice, error) {
	if _, err := s.projectRepo.GetByID(ctx, projectID, userID); err != nil {
		return nil, err
	}
	return s.invoiceRepo.ListByProject(ctx, projectID)
}

func (s *InvoiceService) MarkSent(ctx context.Context, id, userID uuid.UUID) (*domain.Invoice, error) {
	inv, err := s.Get(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	inv.Status = domain.InvoiceStatusSent
	if err := s.invoiceRepo.Update(ctx, inv); err != nil {
		return nil, err
	}
	return inv, nil
}

func (s *InvoiceService) MarkPaid(ctx context.Context, id, userID uuid.UUID) (*domain.Invoice, error) {
	inv, err := s.Get(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	inv.Status = domain.InvoiceStatusPaid
	inv.PaidAt = &now
	if err := s.invoiceRepo.Update(ctx, inv); err != nil {
		return nil, err
	}
	return inv, nil
}

func (s *InvoiceService) CreateShareLink(ctx context.Context, id, userID uuid.UUID, expiresIn time.Duration) (*domain.Invoice, error) {
	inv, err := s.Get(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	shareToken := base64.URLEncoding.EncodeToString(b)
	expiresAt := time.Now().Add(expiresIn).UTC()

	inv.ShareToken = shareToken
	inv.ShareExpiresAt = &expiresAt

	if err := s.invoiceRepo.Update(ctx, inv); err != nil {
		return nil, err
	}
	return inv, nil
}

func (s *InvoiceService) Delete(ctx context.Context, id, userID uuid.UUID) error {
	if _, err := s.Get(ctx, id, userID); err != nil {
		return err
	}
	return s.invoiceRepo.Delete(ctx, id)
}

func generateInvoiceNumber() string {
	now := time.Now()
	b := make([]byte, 3)
	rand.Read(b)
	return fmt.Sprintf("INV-%s-%X", now.Format("200601"), b)
}
