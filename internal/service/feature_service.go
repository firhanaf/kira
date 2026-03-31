package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/kira-app/kira-server/internal/domain"
	"github.com/kira-app/kira-server/internal/repository"
	"github.com/kira-app/kira-server/internal/validator"
)

type FeatureService struct {
	featureRepo *repository.FeatureRepo
	projectRepo *repository.ProjectRepo
}

func NewFeatureService(featureRepo *repository.FeatureRepo, projectRepo *repository.ProjectRepo) *FeatureService {
	return &FeatureService{featureRepo: featureRepo, projectRepo: projectRepo}
}

type CreateFeatureInput struct {
	Name        string
	Description string
	Notes       string
	Severity    string
	GitBranch   string
	GitRepoURL  string
}

func (s *FeatureService) Create(ctx context.Context, projectID, userID uuid.UUID, in CreateFeatureInput) (*domain.Feature, error) {
	if _, err := s.projectRepo.GetByID(ctx, projectID, userID); err != nil {
		return nil, err
	}
	if !validator.NotEmpty(in.Name) {
		return nil, fmt.Errorf("%w: name is required", domain.ErrBadRequest)
	}
	if in.Severity == "" {
		in.Severity = domain.FeatureSeverityMedium
	} else if !validator.OneOf(in.Severity,
		domain.FeatureSeverityCritical, domain.FeatureSeverityHigh,
		domain.FeatureSeverityMedium, domain.FeatureSeverityLow) {
		return nil, fmt.Errorf("%w: invalid severity", domain.ErrBadRequest)
	}

	f := &domain.Feature{
		ID:          uuid.New(),
		ProjectID:   projectID,
		Name:        in.Name,
		Description: in.Description,
		Notes:       in.Notes,
		Severity:    in.Severity,
		Status:      domain.FeatureStatusIdle,
		GitBranch:   in.GitBranch,
		GitRepoURL:  in.GitRepoURL,
	}
	if err := s.featureRepo.Create(ctx, f); err != nil {
		return nil, err
	}
	return f, nil
}

func (s *FeatureService) Get(ctx context.Context, id, userID uuid.UUID) (*domain.Feature, error) {
	f, err := s.featureRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if _, err := s.projectRepo.GetByID(ctx, f.ProjectID, userID); err != nil {
		return nil, domain.ErrForbidden
	}
	return f, nil
}

func (s *FeatureService) ListByProject(ctx context.Context, projectID, userID uuid.UUID) ([]domain.Feature, error) {
	if _, err := s.projectRepo.GetByID(ctx, projectID, userID); err != nil {
		return nil, err
	}
	return s.featureRepo.ListByProject(ctx, projectID)
}

type UpdateFeatureInput struct {
	Name        string
	Description string
	Notes       string
	Severity    string
	Status      string
	GitBranch   string
	GitRepoURL  string
	Position    int
}

func (s *FeatureService) Update(ctx context.Context, id, userID uuid.UUID, in UpdateFeatureInput) (*domain.Feature, error) {
	f, err := s.Get(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if !validator.NotEmpty(in.Name) {
		return nil, fmt.Errorf("%w: name is required", domain.ErrBadRequest)
	}
	if in.Severity != "" && !validator.OneOf(in.Severity,
		domain.FeatureSeverityCritical, domain.FeatureSeverityHigh,
		domain.FeatureSeverityMedium, domain.FeatureSeverityLow) {
		return nil, fmt.Errorf("%w: invalid severity", domain.ErrBadRequest)
	}
	if in.Status != "" && !validator.OneOf(in.Status,
		domain.FeatureStatusIdle, domain.FeatureStatusInProgress,
		domain.FeatureStatusPaused, domain.FeatureStatusDone) {
		return nil, fmt.Errorf("%w: invalid status", domain.ErrBadRequest)
	}

	f.Name = in.Name
	f.Description = in.Description
	f.Notes = in.Notes
	if in.Severity != "" {
		f.Severity = in.Severity
	}
	if in.Status != "" {
		f.Status = in.Status
	}
	f.GitBranch = in.GitBranch
	f.GitRepoURL = in.GitRepoURL
	f.Position = in.Position

	if err := s.featureRepo.Update(ctx, f); err != nil {
		return nil, err
	}
	return f, nil
}

func (s *FeatureService) Delete(ctx context.Context, id, userID uuid.UUID) error {
	if _, err := s.Get(ctx, id, userID); err != nil {
		return err
	}
	return s.featureRepo.Delete(ctx, id)
}
