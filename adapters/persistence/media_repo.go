package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/khoahotran/personal-os/internal/domain/media"
)

type postgresMediaRepo struct {
	db *pgxpool.Pool
}

func NewPostgresMediaRepo(db *pgxpool.Pool) media.Repository {
	return &postgresMediaRepo{db: db}
}

var psqlMedia = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

func scanMedia(row pgx.Row) (*media.Media, error) {
	m := &media.Media{}
	var metadataBytes []byte
	var thumbURL sql.NullString

	err := row.Scan(
		&m.ID, &m.OwnerID, &m.Provider, &m.URL,
		&thumbURL, &m.Status, &metadataBytes,
		&m.IsPublic, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("media not found")
		}
		return nil, fmt.Errorf("failed to scan media row: %w", err)
	}

	if thumbURL.Valid {
		m.ThumbnailURL = &thumbURL.String
	}
	if err := json.Unmarshal(metadataBytes, &m.Metadata); err != nil {
		m.Metadata = map[string]any{}
	}
	return m, nil
}

func scanMedias(rows pgx.Rows) ([]*media.Media, error) {
	defer rows.Close()
	medias := make([]*media.Media, 0)
	for rows.Next() {
		m, err := scanMedia(rows)
		if err != nil {
			return nil, err
		}
		medias = append(medias, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating media rows: %w", err)
	}
	return medias, nil
}

func (r *postgresMediaRepo) Save(ctx context.Context, m *media.Media) error {
	metadataBytes, err := json.Marshal(m.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal media metadata: %w", err)
	}

	query := `
		INSERT INTO media (id, owner_id, provider, url, thumbnail_url, status, metadata, is_public, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err = r.db.Exec(ctx, query,
		m.ID, m.OwnerID, m.Provider, m.URL, m.ThumbnailURL, m.Status,
		metadataBytes, m.IsPublic, m.CreatedAt, m.UpdatedAt,
	)
	return err
}

func (r *postgresMediaRepo) Update(ctx context.Context, m *media.Media) error {
	metadataBytes, err := json.Marshal(m.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal media metadata: %w", err)
	}

	query := `
		UPDATE media SET
			provider = $2, url = $3, thumbnail_url = $4, status = $5, 
			metadata = $6, is_public = $7, updated_at = NOW()
		WHERE id = $1 AND owner_id = $8
	`
	cmdTag, err := r.db.Exec(ctx, query,
		m.ID, m.Provider, m.URL, m.ThumbnailURL, m.Status,
		metadataBytes, m.IsPublic, m.OwnerID,
	)
	if err != nil {
		return fmt.Errorf("failed to update media: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("media not found or user not authorized")
	}
	return nil
}

func (r *postgresMediaRepo) Delete(ctx context.Context, id uuid.UUID, ownerID uuid.UUID) error {
	query := `DELETE FROM media WHERE id = $1 AND owner_id = $2`
	cmdTag, err := r.db.Exec(ctx, query, id, ownerID)
	if err != nil {
		return fmt.Errorf("failed to delete media: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("media not found or user not authorized")
	}
	return nil
}

func (r *postgresMediaRepo) FindByID(ctx context.Context, id uuid.UUID, ownerID uuid.UUID) (*media.Media, error) {
	query := `SELECT * FROM media WHERE id = $1 AND owner_id = $2`
	row := r.db.QueryRow(ctx, query, id, ownerID)
	return scanMedia(row)
}

func (r *postgresMediaRepo) ListPublic(ctx context.Context, limit, offset int) ([]*media.Media, error) {
	builder := psqlMedia.Select("*").
		From("media").
		Where(sq.Eq{"is_public": true, "status": media.StatusReady}).
		OrderBy("created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset))

	sql, args, _ := builder.ToSql()
	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query public media: %w", err)
	}
	return scanMedias(rows)
}

func (r *postgresMediaRepo) ListByOwner(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]*media.Media, error) {
	builder := psqlMedia.Select("*").
		From("media").
		Where(sq.Eq{"owner_id": ownerID}).
		OrderBy("created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset))

	sql, args, _ := builder.ToSql()
	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query media by owner: %w", err)
	}
	return scanMedias(rows)
}
