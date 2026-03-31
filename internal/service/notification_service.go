package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kira-app/kira-server/internal/domain"
	"github.com/kira-app/kira-server/internal/email"
	"github.com/kira-app/kira-server/internal/repository"
)

type NotificationService struct {
	notifRepo   *repository.NotificationRepo
	projectRepo *repository.ProjectRepo
	sender      *email.Sender
	appURL      string
}

func NewNotificationService(
	notifRepo *repository.NotificationRepo,
	projectRepo *repository.ProjectRepo,
	sender *email.Sender,
	appURL string,
) *NotificationService {
	return &NotificationService{
		notifRepo:   notifRepo,
		projectRepo: projectRepo,
		sender:      sender,
		appURL:      appURL,
	}
}

type SendEmailInput struct {
	ProjectID      uuid.UUID
	FeatureID      *uuid.UUID
	RecipientEmail string
	Subject        string
	Message        string
}

func (s *NotificationService) SendEmail(ctx context.Context, userID uuid.UUID, in SendEmailInput) (*domain.Notification, error) {
	if _, err := s.projectRepo.GetByID(ctx, in.ProjectID, userID); err != nil {
		return nil, err
	}

	n := &domain.Notification{
		ID:             uuid.New(),
		ProjectID:      in.ProjectID,
		UserID:         userID,
		FeatureID:      in.FeatureID,
		Channel:        domain.NotificationChannelEmail,
		RecipientEmail: in.RecipientEmail,
		Subject:        in.Subject,
		Message:        in.Message,
		Status:         domain.NotificationStatusPending,
	}

	if err := s.notifRepo.Create(ctx, n); err != nil {
		return nil, err
	}

	htmlBody := fmt.Sprintf("<p>%s</p>", in.Message)
	if err := s.sender.Send(in.RecipientEmail, in.Subject, htmlBody); err != nil {
		_ = s.notifRepo.MarkFailed(ctx, n.ID, err.Error())
		n.Status = domain.NotificationStatusFailed
		return n, nil
	}

	_ = s.notifRepo.MarkSent(ctx, n.ID)
	n.Status = domain.NotificationStatusSent
	return n, nil
}

type CreateShareableLinkInput struct {
	ProjectID uuid.UUID
	FeatureID *uuid.UUID
	Message   string
	ExpiresIn time.Duration
}

func (s *NotificationService) CreateShareableLink(ctx context.Context, userID uuid.UUID, in CreateShareableLinkInput) (*domain.Notification, error) {
	if _, err := s.projectRepo.GetByID(ctx, in.ProjectID, userID); err != nil {
		return nil, err
	}

	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	shareToken := base64.URLEncoding.EncodeToString(b)
	expiresAt := time.Now().Add(in.ExpiresIn).UTC()

	n := &domain.Notification{
		ID:             uuid.New(),
		ProjectID:      in.ProjectID,
		UserID:         userID,
		FeatureID:      in.FeatureID,
		Channel:        domain.NotificationChannelShareableLink,
		Message:        in.Message,
		Status:         domain.NotificationStatusSent,
		ShareToken:     shareToken,
		ShareExpiresAt: &expiresAt,
	}

	if err := s.notifRepo.Create(ctx, n); err != nil {
		return nil, err
	}
	_ = s.notifRepo.MarkSent(ctx, n.ID)

	return n, nil
}

func (s *NotificationService) GetByShareToken(ctx context.Context, token string) (*domain.Notification, error) {
	n, err := s.notifRepo.GetByShareToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if n.ShareExpiresAt != nil && time.Now().After(*n.ShareExpiresAt) {
		return nil, fmt.Errorf("%w: link expired", domain.ErrNotFound)
	}
	return n, nil
}
