package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	ProjectStatusActive    = "active"
	ProjectStatusCompleted = "completed"
	ProjectStatusOnHold    = "on_hold"
	ProjectStatusArchived  = "archived"

	CurrencyIDR = "IDR"
	CurrencyUSD = "USD"
)

type Project struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Name        string
	Description string
	ClientName  string
	ClientEmail string
	Status      string
	Currency    string
	HourlyRate  int // cents/sen per hour
	Color       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}
