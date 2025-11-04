package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/khoahotran/personal-os/internal/domain/post"
	"github.com/khoahotran/personal-os/pkg/apperror"
	"github.com/khoahotran/personal-os/pkg/logger"
	"github.com/pgvector/pgvector-go"
)

type postgresPostRepo struct {
	db     *pgxpool.Pool
	logger logger.Logger
}

func NewPostgresPostRepo(db *pgxpool.Pool, logger logger.Logger) post.Repository {
	return &postgresPostRepo{db: db, logger: logger}
}

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

func scanPost(row pgx.Row, l logger.Logger) (*post.Post, error) {
	p := &post.Post{}
	var historyBytes, metadataBytes []byte
	var ogImageURL, thumbnailURL sql.NullString
	var publishedAt sql.NullTime
	var embedding pgvector.Vector

	err := row.Scan(
		&p.ID,
		&p.OwnerID,
		&p.Slug,
		&p.Title,
		&p.ContentMarkdown,
		&p.Status,
		&ogImageURL,
		&thumbnailURL,
		&metadataBytes,
		&historyBytes,
		&embedding,
		&publishedAt,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, post.ErrPostNotFound
		}
		return nil, apperror.NewInternal("failed to scan post row", err)
	}

	if ogImageURL.Valid {
		p.OgImageURL = &ogImageURL.String
	}

	if thumbnailURL.Valid {
		p.ThumbnailURL = &thumbnailURL.String
	}

	if publishedAt.Valid {
		p.PublishedAt = &publishedAt.Time
	}
	p.Embedding = embedding

	if err := json.Unmarshal(historyBytes, &p.VersionHistory); err != nil {
		l.Warn("Failed to unmarshal post version_history", zap.String("post_id", p.ID.String()), zap.Error(err))
		p.VersionHistory = []post.PostVersion{}
	}
	if err := json.Unmarshal(metadataBytes, &p.Metadata); err != nil {
		l.Warn("Failed to unmarshal post metadata", zap.String("post_id", p.ID.String()), zap.Error(err))
		p.Metadata = map[string]any{}
	}
	return p, nil
}

func scanPosts(rows pgx.Rows, l logger.Logger) ([]*post.Post, error) {
	posts := make([]*post.Post, 0)
	defer rows.Close()

	for rows.Next() {
		p, err := scanPost(rows, l)
		if err != nil {
			return nil, err
		}

		posts = append(posts, p)
	}

	if err := rows.Err(); err != nil {
		return nil, apperror.NewInternal("error iterating post rows", err)
	}
	return posts, nil
}

func (r *postgresPostRepo) Save(ctx context.Context, p *post.Post) error {
	historyBytes, err := json.Marshal(p.VersionHistory)
	if err != nil {
		return apperror.NewInternal("failed to marshal post version_history", err)
	}
	metadataBytes, err := json.Marshal(p.Metadata)
	if err != nil {
		return apperror.NewInternal("failed to marshal post metadata", err)
	}

	if p.Status == post.StatusPublic && p.PublishedAt == nil {
		now := time.Now().UTC()
		p.PublishedAt = &now
	}

	query := `
		INSERT INTO posts (id, owner_id, slug, title, content_markdown, status, metadata, version_history, embedding, published_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err = r.db.Exec(ctx, query,
		p.ID, p.OwnerID, p.Slug, p.Title, p.ContentMarkdown, p.Status,
		metadataBytes, historyBytes, p.Embedding, p.PublishedAt, p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return apperror.NewConflict("post", "slug", p.Slug)
		}
		return apperror.NewInternal("failed to save post", err)
	}
	return nil
}

func (r *postgresPostRepo) Update(ctx context.Context, p *post.Post) error {
	historyBytes, err := json.Marshal(p.VersionHistory)
	if err != nil {
		return apperror.NewInternal("failed to marshal post version_history", err)
	}
	metadataBytes, err := json.Marshal(p.Metadata)
	if err != nil {
		return apperror.NewInternal("failed to marshal post metadata", err)
	}

	if p.Status == post.StatusPublic && p.PublishedAt == nil {
		now := time.Now().UTC()
		p.PublishedAt = &now
	}

	query := `
		UPDATE posts SET
			slug = $2, title = $3, content_markdown = $4, status = $5, 
			version_history = $6, metadata = $7, published_at = $8, og_image_url = $9, thumbnail_url = $10, 
			embedding = $11,
			updated_at = NOW()
		WHERE id = $1 AND owner_id = $12
	`
	cmdTag, err := r.db.Exec(ctx, query,
		p.ID, p.Slug, p.Title, p.ContentMarkdown, p.Status,
		historyBytes, metadataBytes, p.PublishedAt, p.OgImageURL, p.ThumbnailURL, p.Embedding, p.OwnerID,
	)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return apperror.NewConflict("post", "slug", p.Slug)
		}
		return apperror.NewInternal("failed to update post", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return apperror.NewNotFound("post", p.ID.String())
	}
	return nil
}

func (r *postgresPostRepo) Delete(ctx context.Context, id uuid.UUID, ownerID uuid.UUID) error {
	query := `DELETE FROM posts WHERE id = $1 AND owner_id = $2`
	cmdTag, err := r.db.Exec(ctx, query, id, ownerID)
	if err != nil {
		return apperror.NewInternal("failed to delete post", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return apperror.NewNotFound("post", id.String())
	}
	return nil
}

func (r *postgresPostRepo) FindByID(ctx context.Context, id uuid.UUID, ownerID uuid.UUID) (*post.Post, error) {
	query := `SELECT id, owner_id, slug, title, content_markdown, status, og_image_url, thumbnail_url, metadata, version_history, embedding, published_at, created_at, updated_at FROM posts WHERE id = $1 AND owner_id = $2`
	row := r.db.QueryRow(ctx, query, id, ownerID)
	return scanPost(row, r.logger)
}

func (r *postgresPostRepo) FindBySlug(ctx context.Context, slug string) (*post.Post, error) {
	query := `SELECT id, owner_id, slug, title, content_markdown, status, og_image_url, thumbnail_url, metadata, version_history, embedding, published_at, created_at, updated_at FROM posts WHERE slug = $1`
	row := r.db.QueryRow(ctx, query, slug)
	return scanPost(row, r.logger)
}

func (r *postgresPostRepo) FindPublicBySlug(ctx context.Context, slug string) (*post.Post, error) {
	query := `SELECT id, owner_id, slug, title, content_markdown, status, og_image_url, thumbnail_url, metadata, version_history, embedding, published_at, created_at, updated_at FROM posts WHERE slug = $1 AND status = $2`
	row := r.db.QueryRow(ctx, query, slug, post.StatusPublic)
	return scanPost(row, r.logger)
}

func (r *postgresPostRepo) ListByOwner(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]*post.Post, error) {
	builder := psql.Select("*").
		From("posts").
		Where(sq.Eq{"owner_id": ownerID}).
		OrderBy("created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset))

	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, apperror.NewInternal("failed to build list posts by owner query", err)
	}
	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, apperror.NewInternal("failed to query posts by owner", err)
	}
	return scanPosts(rows, r.logger)
}

func (r *postgresPostRepo) ListPublic(ctx context.Context, limit, offset int) ([]*post.Post, error) {
	builder := psql.Select("*").
		From("posts").
		Where(sq.Eq{"status": post.StatusPublic}).
		OrderBy("published_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset))

	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, apperror.NewInternal("failed to build list public posts query", err)
	}
	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, apperror.NewInternal("failed to query public posts", err)
	}
	return scanPosts(rows, r.logger)
}
