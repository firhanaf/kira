package domain

import (
	"time"

	"github.com/google/uuid"
)

type TimeEntry struct {
	ID              uuid.UUID
	FeatureID       uuid.UUID
	UserID          uuid.UUID
	StartedAt       time.Time
	EndedAt         *time.Time
	DurationSeconds *int
	Note            string
	CreatedAt       time.Time
}
