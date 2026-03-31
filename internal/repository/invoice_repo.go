package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kira-app/kira-server/internal/domain"
)

type InvoiceRepo struct {
	db *pgxpool.Pool
}

func NewInvoiceRepo(db *pgxpool.Pool) *InvoiceRepo {
	return &InvoiceRepo{db: db}
}

func (r *InvoiceRepo) Create(ctx context.Context, inv *domain.Invoice) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO invoices (id, project_id, user_id, invoice_number, status,
		                      subtotal, margin_pct, tax_pct, total, currency,
		                      due_date, share_token, share_expires_at, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		inv.ID, inv.ProjectID, inv.UserID, inv.InvoiceNumber, inv.Status,
		inv.Subtotal, inv.MarginPct, inv.TaxPct, inv.Total, inv.Currency,
		inv.DueDate, nullString(inv.ShareToken), inv.ShareExpiresAt, inv.Notes,
	)
	if err != nil {
		return err
	}

	for _, li := range inv.LineItems {
		_, err = tx.Exec(ctx, `
			INSERT INTO invoice_line_items (id, invoice_id, feature_id, feature_name, severity,
			                               git_branch, elapsed_seconds, hourly_rate, amount, notes)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
			li.ID, inv.ID, li.FeatureID, li.FeatureName, li.Severity,
			li.GitBranch, li.ElapsedSeconds, li.HourlyRate, li.Amount, li.Notes,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *InvoiceRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Invoice, error) {
	inv := &domain.Invoice{}
	var shareToken *string
	err := r.db.QueryRow(ctx, `
		SELECT id, project_id, user_id, invoice_number, status,
		       subtotal, margin_pct, tax_pct, total, currency,
		       due_date, paid_at, share_token, share_expires_at, pdf_url, notes, created_at, updated_at
		FROM invoices WHERE id=$1`, id,
	).Scan(&inv.ID, &inv.ProjectID, &inv.UserID, &inv.InvoiceNumber, &inv.Status,
		&inv.Subtotal, &inv.MarginPct, &inv.TaxPct, &inv.Total, &inv.Currency,
		&inv.DueDate, &inv.PaidAt, &shareToken, &inv.ShareExpiresAt, &inv.PDFUrl, &inv.Notes,
		&inv.CreatedAt, &inv.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if shareToken != nil {
		inv.ShareToken = *shareToken
	}

	inv.LineItems, err = r.listLineItems(ctx, id)
	return inv, err
}

func (r *InvoiceRepo) GetByShareToken(ctx context.Context, token string) (*domain.Invoice, error) {
	inv := &domain.Invoice{}
	var shareToken *string
	err := r.db.QueryRow(ctx, `
		SELECT id, project_id, user_id, invoice_number, status,
		       subtotal, margin_pct, tax_pct, total, currency,
		       due_date, paid_at, share_token, share_expires_at, pdf_url, notes, created_at, updated_at
		FROM invoices WHERE share_token=$1`, token,
	).Scan(&inv.ID, &inv.ProjectID, &inv.UserID, &inv.InvoiceNumber, &inv.Status,
		&inv.Subtotal, &inv.MarginPct, &inv.TaxPct, &inv.Total, &inv.Currency,
		&inv.DueDate, &inv.PaidAt, &shareToken, &inv.ShareExpiresAt, &inv.PDFUrl, &inv.Notes,
		&inv.CreatedAt, &inv.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if shareToken != nil {
		inv.ShareToken = *shareToken
	}

	inv.LineItems, err = r.listLineItems(ctx, inv.ID)
	return inv, err
}

func (r *InvoiceRepo) ListByProject(ctx context.Context, projectID uuid.UUID) ([]domain.Invoice, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, project_id, user_id, invoice_number, status,
		       subtotal, margin_pct, tax_pct, total, currency,
		       due_date, paid_at, share_token, share_expires_at, pdf_url, notes, created_at, updated_at
		FROM invoices WHERE project_id=$1 ORDER BY created_at DESC`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invoices []domain.Invoice
	for rows.Next() {
		var inv domain.Invoice
		var shareToken *string
		if err := rows.Scan(&inv.ID, &inv.ProjectID, &inv.UserID, &inv.InvoiceNumber, &inv.Status,
			&inv.Subtotal, &inv.MarginPct, &inv.TaxPct, &inv.Total, &inv.Currency,
			&inv.DueDate, &inv.PaidAt, &shareToken, &inv.ShareExpiresAt, &inv.PDFUrl, &inv.Notes,
			&inv.CreatedAt, &inv.UpdatedAt); err != nil {
			return nil, err
		}
		if shareToken != nil {
			inv.ShareToken = *shareToken
		}
		invoices = append(invoices, inv)
	}
	return invoices, rows.Err()
}

func (r *InvoiceRepo) Update(ctx context.Context, inv *domain.Invoice) error {
	_, err := r.db.Exec(ctx, `
		UPDATE invoices SET status=$1, due_date=$2, paid_at=$3, share_token=$4,
		       share_expires_at=$5, pdf_url=$6, notes=$7, margin_pct=$8, tax_pct=$9, total=$10
		WHERE id=$11`,
		inv.Status, inv.DueDate, inv.PaidAt, nullString(inv.ShareToken),
		inv.ShareExpiresAt, inv.PDFUrl, inv.Notes, inv.MarginPct, inv.TaxPct, inv.Total, inv.ID,
	)
	return err
}

func (r *InvoiceRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM invoices WHERE id=$1`, id)
	return err
}

func (r *InvoiceRepo) listLineItems(ctx context.Context, invoiceID uuid.UUID) ([]domain.InvoiceLineItem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, invoice_id, feature_id, feature_name, severity,
		       git_branch, elapsed_seconds, hourly_rate, amount, notes
		FROM invoice_line_items WHERE invoice_id=$1`, invoiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.InvoiceLineItem
	for rows.Next() {
		var li domain.InvoiceLineItem
		if err := rows.Scan(&li.ID, &li.InvoiceID, &li.FeatureID, &li.FeatureName, &li.Severity,
			&li.GitBranch, &li.ElapsedSeconds, &li.HourlyRate, &li.Amount, &li.Notes); err != nil {
			return nil, err
		}
		items = append(items, li)
	}
	return items, rows.Err()
}

// ─── Notification Repo ───────────────────────────────────────────────────────

type NotificationRepo struct {
	db *pgxpool.Pool
}

func NewNotificationRepo(db *pgxpool.Pool) *NotificationRepo {
	return &NotificationRepo{db: db}
}

func (r *NotificationRepo) Create(ctx context.Context, n *domain.Notification) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO notifications (id, project_id, user_id, feature_id, channel,
		                           recipient_email, subject, message, status, share_token, share_expires_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		n.ID, n.ProjectID, n.UserID, n.FeatureID, n.Channel,
		n.RecipientEmail, n.Subject, n.Message, n.Status, nullString(n.ShareToken), n.ShareExpiresAt,
	)
	return err
}

func (r *NotificationRepo) MarkSent(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`UPDATE notifications SET status='sent', sent_at=NOW() WHERE id=$1`, id)
	return err
}

func (r *NotificationRepo) MarkFailed(ctx context.Context, id uuid.UUID, errMsg string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE notifications SET status='failed', error_message=$1 WHERE id=$2`, errMsg, id)
	return err
}

func (r *NotificationRepo) GetByShareToken(ctx context.Context, token string) (*domain.Notification, error) {
	n := &domain.Notification{}
	var shareToken *string
	err := r.db.QueryRow(ctx, `
		SELECT id, project_id, user_id, feature_id, channel,
		       recipient_email, subject, message, status, share_token, share_expires_at, sent_at, created_at
		FROM notifications WHERE share_token=$1`, token,
	).Scan(&n.ID, &n.ProjectID, &n.UserID, &n.FeatureID, &n.Channel,
		&n.RecipientEmail, &n.Subject, &n.Message, &n.Status, &shareToken, &n.ShareExpiresAt,
		&n.SentAt, &n.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if shareToken != nil {
		n.ShareToken = *shareToken
	}
	return n, nil
}

// nullString converts empty string to nil for nullable TEXT columns.
func nullString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
