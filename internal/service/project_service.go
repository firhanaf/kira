package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/kira-app/kira-server/internal/domain"
	"github.com/kira-app/kira-server/internal/repository"
	"github.com/kira-app/kira-server/internal/validator"
)

type ProjectService struct {
	repo *repository.ProjectRepo
}

func NewProjectService(repo *repository.ProjectRepo) *ProjectService {
	return &ProjectService{repo: repo}
}

type CreateProjectInput struct {
	Name        string
	Description string
	ClientName  string
	ClientEmail string
	Currency    string
	HourlyRate  int
	Color       string
}

func (s *ProjectService) Create(ctx context.Context, userID uuid.UUID, in CreateProjectInput) (*domain.Project, error) {
	if !validator.NotEmpty(in.Name) {
		return nil, fmt.Errorf("%w: name is required", domain.ErrBadRequest)
	}
	if !validator.OneOf(in.Currency, domain.CurrencyIDR, domain.CurrencyUSD) {
		in.Currency = domain.CurrencyIDR
	}
	if in.Color == "" {
		in.Color = "#6366f1"
	}

	p := &domain.Project{
		ID:          uuid.New(),
		UserID:      userID,
		Name:        in.Name,
		Description: in.Description,
		ClientName:  in.ClientName,
		ClientEmail: in.ClientEmail,
		Status:      domain.ProjectStatusActive,
		Currency:    in.Currency,
		HourlyRate:  in.HourlyRate,
		Color:       in.Color,
	}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *ProjectService) Get(ctx context.Context, id, userID uuid.UUID) (*domain.Project, error) {
	return s.repo.GetByID(ctx, id, userID)
}

func (s *ProjectService) List(ctx context.Context, userID uuid.UUID) ([]domain.Project, error) {
	return s.repo.List(ctx, userID)
}

type UpdateProjectInput struct {
	Name        string
	Description string
	ClientName  string
	ClientEmail string
	Status      string
	Currency    string
	HourlyRate  int
	Color       string
}

func (s *ProjectService) Update(ctx context.Context, id, userID uuid.UUID, in UpdateProjectInput) (*domain.Project, error) {
	p, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if !validator.NotEmpty(in.Name) {
		return nil, fmt.Errorf("%w: name is required", domain.ErrBadRequest)
	}
	if in.Status != "" && !validator.OneOf(in.Status,
		domain.ProjectStatusActive, domain.ProjectStatusCompleted,
		domain.ProjectStatusOnHold, domain.ProjectStatusArchived) {
		return nil, fmt.Errorf("%w: invalid status", domain.ErrBadRequest)
	}

	p.Name = in.Name
	p.Description = in.Description
	p.ClientName = in.ClientName
	p.ClientEmail = in.ClientEmail
	if in.Status != "" {
		p.Status = in.Status
	}
	if validator.OneOf(in.Currency, domain.CurrencyIDR, domain.CurrencyUSD) {
		p.Currency = in.Currency
	}
	p.HourlyRate = in.HourlyRate
	if in.Color != "" {
		p.Color = in.Color
	}

	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *ProjectService) Delete(ctx context.Context, id, userID uuid.UUID) error {
	if _, err := s.repo.GetByID(ctx, id, userID); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id, userID)
}
