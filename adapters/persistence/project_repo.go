package persistence

import (
	"context"
	"encoding/json"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/khoahotran/personal-os/internal/domain/project"
	"github.com/khoahotran/personal-os/pkg/apperror"
	"github.com/khoahotran/personal-os/pkg/logger"
)

type postgresProjectRepo struct {
	db     *pgxpool.Pool
	logger logger.Logger
}

func NewPostgresProjectRepo(db *pgxpool.Pool, logger logger.Logger) project.Repository {
	return &postgresProjectRepo{db: db, logger: logger}
}

var psqlProject = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

func scanProject(row pgx.Row, l logger.Logger) (*project.Project, error) {
	p := &project.Project{}
	var mediaBytes []byte

	err := row.Scan(
		&p.ID,
		&p.OwnerID,
		&p.Slug,
		&p.Title,
		&p.Description,
		&p.Stack,
		&p.RepositoryURL,
		&p.LiveURL,
		&mediaBytes,
		&p.IsPublic,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.NewNotFound("project", "")
		}
		return nil, apperror.NewInternal("failed to scan project row", err)
	}

	if err := json.Unmarshal(mediaBytes, &p.Media); err != nil {
		l.Warn("Failed to unmarshal project media", zap.String("project_id", p.ID.String()), zap.Error(err))
		p.Media = []project.ProjectMedia{}
	}

	return p, nil
}

func scanProjects(rows pgx.Rows, l logger.Logger) ([]*project.Project, error) {
	defer rows.Close()
	projects := make([]*project.Project, 0)

	for rows.Next() {
		p, err := scanProject(rows, l)
		if err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	if err := rows.Err(); err != nil {
		return nil, apperror.NewInternal("error iterating project rows", err)
	}
	return projects, nil
}

func (r *postgresProjectRepo) Save(ctx context.Context, p *project.Project) error {
	mediaBytes, err := json.Marshal(p.Media)
	if err != nil {
		return apperror.NewInternal("failed to marshal project media", err)
	}

	query := `
		INSERT INTO projects (id, owner_id, slug, title, description, stack, repository_url, live_url, media, is_public, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err = r.db.Exec(ctx, query,
		p.ID, p.OwnerID, p.Slug, p.Title, p.Description,
		(p.Stack), p.RepositoryURL, p.LiveURL, mediaBytes,
		p.IsPublic, p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return apperror.NewConflict("project", "slug", p.Slug)
		}
		return apperror.NewInternal("failed to save project", err)
	}
	return nil
}

func (r *postgresProjectRepo) Update(ctx context.Context, p *project.Project) error {
	mediaBytes, err := json.Marshal(p.Media)
	if err != nil {
		return apperror.NewInternal("failed to marshal project media for update", err)
	}

	query := `
		UPDATE projects SET
			slug = $2, title = $3, description = $4, stack = $5, repository_url = $6,
			live_url = $7, media = $8, is_public = $9, updated_at = NOW()
		WHERE id = $1 AND owner_id = $10
	`
	cmdTag, err := r.db.Exec(ctx, query,
		p.ID, p.Slug, p.Title, p.Description, p.Stack,
		p.RepositoryURL, p.LiveURL, mediaBytes, p.IsPublic,
		p.OwnerID,
	)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return apperror.NewConflict("project", "slug", p.Slug)
		}
		return apperror.NewInternal("failed to update project", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return apperror.NewNotFound("project", p.ID.String())
	}
	return nil
}

func (r *postgresProjectRepo) Delete(ctx context.Context, id uuid.UUID, ownerID uuid.UUID) error {
	query := `DELETE FROM projects WHERE id = $1 AND owner_id = $2`
	cmdTag, err := r.db.Exec(ctx, query, id, ownerID)
	if err != nil {
		return apperror.NewInternal("failed to delete project", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return apperror.NewNotFound("project", id.String())
	}
	return nil
}

func (r *postgresProjectRepo) FindByID(ctx context.Context, id uuid.UUID, ownerID uuid.UUID) (*project.Project, error) {
	query := `
		SELECT id, owner_id, slug, title, description, stack, repository_url, live_url, media, is_public, created_at, updated_at
		FROM projects
		WHERE id = $1 AND owner_id = $2
	`
	row := r.db.QueryRow(ctx, query, id, ownerID)
	return scanProject(row, r.logger)
}

func (r *postgresProjectRepo) FindBySlug(ctx context.Context, slug string) (*project.Project, error) {
	query := `
		SELECT id, owner_id, slug, title, description, stack, repository_url, live_url, media, is_public, created_at, updated_at
		FROM projects
		WHERE slug = $1
	`
	row := r.db.QueryRow(ctx, query, slug)
	return scanProject(row, r.logger)
}

func (r *postgresProjectRepo) FindPublicBySlug(ctx context.Context, slug string) (*project.Project, error) {
	query := `
		SELECT id, owner_id, slug, title, description, stack, repository_url, live_url, media, is_public, created_at, updated_at
		FROM projects
		WHERE slug = $1 AND is_public = true
	`
	row := r.db.QueryRow(ctx, query, slug)
	return scanProject(row, r.logger)
}

func (r *postgresProjectRepo) ListByOwner(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]*project.Project, error) {
	builder := psqlProject.Select("id, owner_id, slug, title, description, stack, repository_url, live_url, media, is_public, created_at, updated_at").
		From("projects").
		Where(sq.Eq{"owner_id": ownerID}).
		OrderBy("created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset))

	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, apperror.NewInternal("failed to build find by owner query", err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, apperror.NewInternal("failed to query projects by owner", err)
	}

	return scanProjects(rows, r.logger)
}

func (r *postgresProjectRepo) ListPublic(ctx context.Context, limit, offset int) ([]*project.Project, error) {
	builder := psqlProject.Select("id, owner_id, slug, title, description, stack, repository_url, live_url, media, is_public, created_at, updated_at").
		From("projects").
		Where(sq.Eq{"is_public": true}).
		OrderBy("created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset))

	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, apperror.NewInternal("failed to build find public projects query", err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, apperror.NewInternal("failed to query public projects", err)
	}

	return scanProjects(rows, r.logger)
}
