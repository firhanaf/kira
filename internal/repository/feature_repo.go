package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kira-app/kira-server/internal/domain"
)

type FeatureRepo struct {
	db *pgxpool.Pool
}

func NewFeatureRepo(db *pgxpool.Pool) *FeatureRepo {
	return &FeatureRepo{db: db}
}

func (r *FeatureRepo) Create(ctx context.Context, f *domain.Feature) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO features (id, project_id, name, description, notes, severity, status,
		                      git_branch, git_repo_url, elapsed_seconds, position)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		f.ID, f.ProjectID, f.Name, f.Description, f.Notes, f.Severity, f.Status,
		f.GitBranch, f.GitRepoURL, f.ElapsedSeconds, f.Position,
	)
	return err
}

func (r *FeatureRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Feature, error) {
	f := &domain.Feature{}
	err := r.db.QueryRow(ctx, `
		SELECT id, project_id, name, description, notes, severity, status,
		       git_branch, git_repo_url, elapsed_seconds, position, created_at, updated_at, deleted_at
		FROM features WHERE id=$1 AND deleted_at IS NULL`, id,
	).Scan(&f.ID, &f.ProjectID, &f.Name, &f.Description, &f.Notes, &f.Severity, &f.Status,
		&f.GitBranch, &f.GitRepoURL, &f.ElapsedSeconds, &f.Position,
		&f.CreatedAt, &f.UpdatedAt, &f.DeletedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return f, err
}

func (r *FeatureRepo) ListByProject(ctx context.Context, projectID uuid.UUID) ([]domain.Feature, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, project_id, name, description, notes, severity, status,
		       git_branch, git_repo_url, elapsed_seconds, position, created_at, updated_at, deleted_at
		FROM features WHERE project_id=$1 AND deleted_at IS NULL ORDER BY position, created_at`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var features []domain.Feature
	for rows.Next() {
		var f domain.Feature
		if err := rows.Scan(&f.ID, &f.ProjectID, &f.Name, &f.Description, &f.Notes, &f.Severity, &f.Status,
			&f.GitBranch, &f.GitRepoURL, &f.ElapsedSeconds, &f.Position,
			&f.CreatedAt, &f.UpdatedAt, &f.DeletedAt); err != nil {
			return nil, err
		}
		features = append(features, f)
	}
	return features, rows.Err()
}

func (r *FeatureRepo) Update(ctx context.Context, f *domain.Feature) error {
	_, err := r.db.Exec(ctx, `
		UPDATE features SET name=$1, description=$2, notes=$3, severity=$4, status=$5,
		       git_branch=$6, git_repo_url=$7, position=$8
		WHERE id=$9 AND deleted_at IS NULL`,
		f.Name, f.Description, f.Notes, f.Severity, f.Status,
		f.GitBranch, f.GitRepoURL, f.Position, f.ID,
	)
	return err
}

func (r *FeatureRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`UPDATE features SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	return err
}

func (r *FeatureRepo) AddElapsed(ctx context.Context, id uuid.UUID, seconds int) error {
	_, err := r.db.Exec(ctx,
		`UPDATE features SET elapsed_seconds = elapsed_seconds + $1 WHERE id=$2`, seconds, id)
	return err
}
