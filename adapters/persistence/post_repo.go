package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/khoahotran/personal-os/internal/domain/post"
)

type postgresPostRepo struct {
	db *pgxpool.Pool
}

func NewPostgresPostRepo(db *pgxpool.Pool) post.Repository {
	return &postgresPostRepo{db: db}
}

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

func scanPost(row pgx.Row) (*post.Post, error) {
	p := &post.Post{}
	var historyBytes []byte
	var ogImageURL, thumbnailURL, publishedAt sql.NullString

	err := row.Scan(
		&p.ID,
		&p.OwnerID,
		&p.Slug,
		&p.Title,
		&p.ContentMarkdown,
		&p.Status,
		&ogImageURL,
		&thumbnailURL,
		&p.Metadata,
		&historyBytes,
		&publishedAt,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, post.ErrPostNotFound
		}
		return nil, fmt.Errorf("failed to scan post row: %w", err)
	}

	if ogImageURL.Valid {
		p.OgImageURL = &ogImageURL.String
	}

	if thumbnailURL.Valid {
		p.ThumbnailURL = &thumbnailURL.String
	}

	if publishedAt.Valid {
		if t, err := time.Parse(time.RFC3339, publishedAt.String); err == nil {
			p.PublishedAt = &t
		}
	}

	if err := json.Unmarshal(historyBytes, &p.VersionHistory); err != nil {
		p.VersionHistory = []post.PostVersion{}
	}
	return p, nil
}

func scanPosts(rows pgx.Rows) ([]*post.Post, error) {
	posts := make([]*post.Post, 0)
	defer rows.Close()

	for rows.Next() {
		p := &post.Post{}
		var historyBytes []byte
		var ogImageURL, thumbnailURL, publishedAt sql.NullString

		err := rows.Scan(
			&p.ID,
			&p.OwnerID,
			&p.Slug,
			&p.Title,
			&p.ContentMarkdown,
			&p.Status,
			&ogImageURL,
			&thumbnailURL,
			&p.Metadata,
			&historyBytes,
			&publishedAt,
			&p.CreatedAt,
			&p.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan post row during iteration: %w", err)
		}

		if ogImageURL.Valid {
			p.OgImageURL = &ogImageURL.String
		}
		if thumbnailURL.Valid {
			p.ThumbnailURL = &thumbnailURL.String
		}
		if publishedAt.Valid {
			if t, err := time.Parse(time.RFC3339, publishedAt.String); err == nil {
				p.PublishedAt = &t
			}
		}
		if err := json.Unmarshal(historyBytes, &p.VersionHistory); err != nil {
			p.VersionHistory = []post.PostVersion{}
		}
		posts = append(posts, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating post rows: %w", err)
	}
	return posts, nil
}

func (r *postgresPostRepo) Save(ctx context.Context, p *post.Post) error {
	historyBytes, err := json.Marshal(p.VersionHistory)
	if err != nil {
		return fmt.Errorf("failed to marshal post version_history: %w", err)
	}
	metadataBytes, err := json.Marshal(p.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal post metadata: %w", err)
	}

	if p.Status == post.StatusPublic && p.PublishedAt == nil {
		now := time.Now().UTC()
		p.PublishedAt = &now
	}

	query := `
		INSERT INTO posts (id, owner_id, slug, title, content_markdown, status, version_history, metadata, published_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err = r.db.Exec(ctx, query,
		p.ID, p.OwnerID, p.Slug, p.Title, p.ContentMarkdown, p.Status,
		historyBytes, metadataBytes, p.PublishedAt, p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return errors.New("post slug already exists")
		}
		return fmt.Errorf("failed to save post: %w", err)
	}
	return nil
}

func (r *postgresPostRepo) Update(ctx context.Context, p *post.Post) error {
	historyBytes, err := json.Marshal(p.VersionHistory)
	if err != nil {
		return fmt.Errorf("failed to marshal post version_history for update: %w", err)
	}

	if p.Status == post.StatusPublic && p.PublishedAt == nil {
		now := time.Now().UTC()
		p.PublishedAt = &now
	}

	// no update metadata
	query := `
		UPDATE posts SET
			slug = $2, title = $3, content_markdown = $4, status = $5, 
			version_history = $6, published_at = $7, og_image_url = $8, thumbnail_url = $9, metadata = $10, updated_at = NOW()
		WHERE id = $1 AND owner_id = $11
	`
	cmdTag, err := r.db.Exec(ctx, query,
		p.ID, p.Slug, p.Title, p.ContentMarkdown, p.Status,
		historyBytes, p.PublishedAt, p.OgImageURL, p.ThumbnailURL, p.Metadata, p.OwnerID,
	)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return errors.New("post slug already exists")
		}
		return fmt.Errorf("failed to update post: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("post not found or user not authorized to update")
	}
	return nil
}

func (r *postgresPostRepo) Delete(ctx context.Context, id uuid.UUID, ownerID uuid.UUID) error {
	query := `DELETE FROM posts WHERE id = $1 AND owner_id = $2`
	cmdTag, err := r.db.Exec(ctx, query, id, ownerID)
	if err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("post not found or user not authorized to delete")
	}
	return nil
}

func (r *postgresPostRepo) FindByID(ctx context.Context, id uuid.UUID, ownerID uuid.UUID) (*post.Post, error) {
	query := `SELECT * FROM posts WHERE id = $1 AND owner_id = $2`
	row := r.db.QueryRow(ctx, query, id, ownerID)
	return scanPost(row)
}

func (r *postgresPostRepo) FindBySlug(ctx context.Context, slug string) (*post.Post, error) {
	query := `SELECT * FROM posts WHERE slug = $1`
	row := r.db.QueryRow(ctx, query, slug)
	return scanPost(row)
}

func (r *postgresPostRepo) FindPublicBySlug(ctx context.Context, slug string) (*post.Post, error) {
	query := `SELECT * FROM posts WHERE slug = $1 AND status = $2`
	row := r.db.QueryRow(ctx, query, slug, post.StatusPublic)
	return scanPost(row)
}

func (r *postgresPostRepo) ListByOwner(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]*post.Post, error) {
	builder := psql.Select("*").
		From("posts").
		Where(sq.Eq{"owner_id": ownerID}).
		OrderBy("created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset))

	sql, args, _ := builder.ToSql()
	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query posts by owner: %w", err)
	}
	return scanPosts(rows)
}

func (r *postgresPostRepo) ListPublic(ctx context.Context, limit, offset int) ([]*post.Post, error) {
	builder := psql.Select("*").
		From("posts").
		Where(sq.Eq{"status": post.StatusPublic}).
		OrderBy("published_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset))

	sql, args, _ := builder.ToSql()
	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query public posts: %w", err)
	}
	return scanPosts(rows)
}
