package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kira-app/kira-server/internal/domain"
)

type ProjectRepo struct {
	db *pgxpool.Pool
}

func NewProjectRepo(db *pgxpool.Pool) *ProjectRepo {
	return &ProjectRepo{db: db}
}

func (r *ProjectRepo) Create(ctx context.Context, p *domain.Project) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO projects (id, user_id, name, description, client_name, client_email,
		                      status, currency, hourly_rate, color)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		p.ID, p.UserID, p.Name, p.Description, p.ClientName, p.ClientEmail,
		p.Status, p.Currency, p.HourlyRate, p.Color,
	)
	return err
}

func (r *ProjectRepo) GetByID(ctx context.Context, id, userID uuid.UUID) (*domain.Project, error) {
	p := &domain.Project{}
	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, name, description, client_name, client_email,
		       status, currency, hourly_rate, color, created_at, updated_at, deleted_at
		FROM projects WHERE id=$1 AND user_id=$2 AND deleted_at IS NULL`, id, userID,
	).Scan(&p.ID, &p.UserID, &p.Name, &p.Description, &p.ClientName, &p.ClientEmail,
		&p.Status, &p.Currency, &p.HourlyRate, &p.Color,
		&p.CreatedAt, &p.UpdatedAt, &p.DeletedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return p, err
}

func (r *ProjectRepo) List(ctx context.Context, userID uuid.UUID) ([]domain.Project, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, name, description, client_name, client_email,
		       status, currency, hourly_rate, color, created_at, updated_at, deleted_at
		FROM projects WHERE user_id=$1 AND deleted_at IS NULL ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []domain.Project
	for rows.Next() {
		var p domain.Project
		if err := rows.Scan(&p.ID, &p.UserID, &p.Name, &p.Description, &p.ClientName, &p.ClientEmail,
			&p.Status, &p.Currency, &p.HourlyRate, &p.Color,
			&p.CreatedAt, &p.UpdatedAt, &p.DeletedAt); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

func (r *ProjectRepo) Update(ctx context.Context, p *domain.Project) error {
	_, err := r.db.Exec(ctx, `
		UPDATE projects SET name=$1, description=$2, client_name=$3, client_email=$4,
		       status=$5, currency=$6, hourly_rate=$7, color=$8
		WHERE id=$9 AND user_id=$10 AND deleted_at IS NULL`,
		p.Name, p.Description, p.ClientName, p.ClientEmail,
		p.Status, p.Currency, p.HourlyRate, p.Color, p.ID, p.UserID,
	)
	return err
}

func (r *ProjectRepo) Delete(ctx context.Context, id, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`UPDATE projects SET deleted_at=NOW() WHERE id=$1 AND user_id=$2 AND deleted_at IS NULL`,
		id, userID)
	return err
}
