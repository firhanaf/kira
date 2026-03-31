package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kira-app/kira-server/internal/domain"
	"github.com/kira-app/kira-server/internal/repository"
	"github.com/kira-app/kira-server/internal/sse"
)

type TimerService struct {
	timerRepo   *repository.TimerRepo
	featureRepo *repository.FeatureRepo
	projectRepo *repository.ProjectRepo
	hub         *sse.Hub
}

func NewTimerService(
	timerRepo *repository.TimerRepo,
	featureRepo *repository.FeatureRepo,
	projectRepo *repository.ProjectRepo,
	hub *sse.Hub,
) *TimerService {
	return &TimerService{
		timerRepo:   timerRepo,
		featureRepo: featureRepo,
		projectRepo: projectRepo,
		hub:         hub,
	}
}

func (s *TimerService) Start(ctx context.Context, featureID, userID uuid.UUID, note string) (*domain.TimeEntry, error) {
	// verify feature belongs to a project owned by user
	f, err := s.featureRepo.GetByID(ctx, featureID)
	if err != nil {
		return nil, err
	}
	if _, err := s.projectRepo.GetByID(ctx, f.ProjectID, userID); err != nil {
		return nil, domain.ErrForbidden
	}

	// check no active timer
	active, err := s.timerRepo.GetActiveByUser(ctx, userID)
	if err != nil && err != domain.ErrNotFound {
		return nil, err
	}
	if active != nil {
		return nil, fmt.Errorf("%w: a timer is already running", domain.ErrConflict)
	}

	entry := &domain.TimeEntry{
		ID:        uuid.New(),
		FeatureID: featureID,
		UserID:    userID,
		StartedAt: time.Now().UTC(),
		Note:      note,
	}
	if err := s.timerRepo.Create(ctx, entry); err != nil {
		return nil, err
	}

	// update feature status
	f.Status = domain.FeatureStatusInProgress
	_ = s.featureRepo.Update(ctx, f)

	s.hub.PublishJSON(userID, "timer.started", entry)
	return entry, nil
}

func (s *TimerService) Stop(ctx context.Context, userID uuid.UUID) (*domain.TimeEntry, error) {
	entry, err := s.timerRepo.GetActiveByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%w: no active timer", domain.ErrNotFound)
	}

	endedAt := time.Now().UTC()
	duration := int(endedAt.Sub(entry.StartedAt).Seconds())

	if err := s.timerRepo.Stop(ctx, entry.ID, endedAt, duration); err != nil {
		return nil, err
	}
	if err := s.featureRepo.AddElapsed(ctx, entry.FeatureID, duration); err != nil {
		return nil, err
	}

	// update feature status to paused
	f, err := s.featureRepo.GetByID(ctx, entry.FeatureID)
	if err == nil {
		f.Status = domain.FeatureStatusPaused
		_ = s.featureRepo.Update(ctx, f)
	}

	entry.EndedAt = &endedAt
	entry.DurationSeconds = &duration

	s.hub.PublishJSON(userID, "timer.stopped", entry)
	return entry, nil
}

func (s *TimerService) GetActive(ctx context.Context, userID uuid.UUID) (*domain.TimeEntry, error) {
	return s.timerRepo.GetActiveByUser(ctx, userID)
}

func (s *TimerService) ListByFeature(ctx context.Context, featureID, userID uuid.UUID) ([]domain.TimeEntry, error) {
	f, err := s.featureRepo.GetByID(ctx, featureID)
	if err != nil {
		return nil, err
	}
	if _, err := s.projectRepo.GetByID(ctx, f.ProjectID, userID); err != nil {
		return nil, domain.ErrForbidden
	}
	return s.timerRepo.ListByFeature(ctx, featureID)
}
