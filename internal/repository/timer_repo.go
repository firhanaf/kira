package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kira-app/kira-server/internal/domain"
)

type TimerRepo struct {
	db *pgxpool.Pool
}

func NewTimerRepo(db *pgxpool.Pool) *TimerRepo {
	return &TimerRepo{db: db}
}

func (r *TimerRepo) Create(ctx context.Context, e *domain.TimeEntry) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO time_entries (id, feature_id, user_id, started_at, note)
		VALUES ($1, $2, $3, $4, $5)`,
		e.ID, e.FeatureID, e.UserID, e.StartedAt, e.Note,
	)
	return err
}

func (r *TimerRepo) GetActiveByUser(ctx context.Context, userID uuid.UUID) (*domain.TimeEntry, error) {
	e := &domain.TimeEntry{}
	err := r.db.QueryRow(ctx, `
		SELECT id, feature_id, user_id, started_at, ended_at, duration_seconds, note, created_at
		FROM time_entries WHERE user_id=$1 AND ended_at IS NULL`, userID,
	).Scan(&e.ID, &e.FeatureID, &e.UserID, &e.StartedAt, &e.EndedAt, &e.DurationSeconds, &e.Note, &e.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return e, err
}

func (r *TimerRepo) Stop(ctx context.Context, id uuid.UUID, endedAt time.Time, durationSeconds int) error {
	_, err := r.db.Exec(ctx, `
		UPDATE time_entries SET ended_at=$1, duration_seconds=$2 WHERE id=$3`,
		endedAt, durationSeconds, id,
	)
	return err
}

func (r *TimerRepo) ListByFeature(ctx context.Context, featureID uuid.UUID) ([]domain.TimeEntry, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, feature_id, user_id, started_at, ended_at, duration_seconds, note, created_at
		FROM time_entries WHERE feature_id=$1 ORDER BY started_at DESC`, featureID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []domain.TimeEntry
	for rows.Next() {
		var e domain.TimeEntry
		if err := rows.Scan(&e.ID, &e.FeatureID, &e.UserID, &e.StartedAt, &e.EndedAt,
			&e.DurationSeconds, &e.Note, &e.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}
