package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	InvoiceStatusDraft   = "draft"
	InvoiceStatusSent    = "sent"
	InvoiceStatusPaid    = "paid"
	InvoiceStatusOverdue = "overdue"

	NotificationChannelEmail         = "email"
	NotificationChannelShareableLink = "shareable_link"

	NotificationStatusPending = "pending"
	NotificationStatusSent    = "sent"
	NotificationStatusFailed  = "failed"
)

type Invoice struct {
	ID             uuid.UUID
	ProjectID      uuid.UUID
	UserID         uuid.UUID
	InvoiceNumber  string
	Status         string
	Subtotal       int // cents/sen
	MarginPct      float64
	TaxPct         float64
	Total          int // cents/sen
	Currency       string
	DueDate        *time.Time
	PaidAt         *time.Time
	ShareToken     string
	ShareExpiresAt *time.Time
	PDFUrl         string
	Notes          string
	LineItems      []InvoiceLineItem
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type InvoiceLineItem struct {
	ID             uuid.UUID
	InvoiceID      uuid.UUID
	FeatureID      *uuid.UUID
	FeatureName    string
	Severity       string
	GitBranch      string
	ElapsedSeconds int
	HourlyRate     int // cents/sen per hour
	Amount         int // cents/sen
	Notes          string
}

type Notification struct {
	ID             uuid.UUID
	ProjectID      uuid.UUID
	UserID         uuid.UUID
	FeatureID      *uuid.UUID
	Channel        string
	RecipientEmail string
	Subject        string
	Message        string
	Status         string
	ShareToken     string
	ShareExpiresAt *time.Time
	SentAt         *time.Time
	ErrorMessage   string
	CreatedAt      time.Time
}
