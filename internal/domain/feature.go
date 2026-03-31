package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	FeatureSeverityCritical = "critical"
	FeatureSeverityHigh     = "high"
	FeatureSeverityMedium   = "medium"
	FeatureSeverityLow      = "low"

	FeatureStatusIdle       = "idle"
	FeatureStatusInProgress = "in_progress"
	FeatureStatusPaused     = "paused"
	FeatureStatusDone       = "done"
)

type Feature struct {
	ID             uuid.UUID
	ProjectID      uuid.UUID
	Name           string
	Description    string
	Notes          string
	Severity       string
	Status         string
	GitBranch      string
	GitRepoURL     string
	ElapsedSeconds int
	Position       int
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time
}
