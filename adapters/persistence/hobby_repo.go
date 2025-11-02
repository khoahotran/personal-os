package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/khoahotran/personal-os/internal/domain/hobby"
)

type postgresHobbyRepo struct {
	db *pgxpool.Pool
}

func NewPostgresHobbyRepo(db *pgxpool.Pool) hobby.Repository {
	return &postgresHobbyRepo{db: db}
}

var psqlHobby = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

func scanHobbyItem(row pgx.Row) (*hobby.HobbyItem, error) {
	hi := &hobby.HobbyItem{}
	var metadataBytes []byte

	err := row.Scan(
		&hi.ID, &hi.OwnerID, &hi.Category, &hi.Title,
		&hi.Status, &hi.Rating, &hi.Notes, &metadataBytes,
		&hi.IsPublic, &hi.CreatedAt, &hi.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, hobby.ErrHobbyItemNotFound
		}
		return nil, fmt.Errorf("failed to scan hobby item row: %w", err)
	}

	if err := json.Unmarshal(metadataBytes, &hi.Metadata); err != nil {
		hi.Metadata = map[string]any{}
	}
	return hi, nil
}

func scanHobbyItems(rows pgx.Rows) ([]*hobby.HobbyItem, error) {
	defer rows.Close()
	items := make([]*hobby.HobbyItem, 0)
	for rows.Next() {

		hi := &hobby.HobbyItem{}
		var metadataBytes []byte
		err := rows.Scan(
			&hi.ID, &hi.OwnerID, &hi.Category, &hi.Title,
			&hi.Status, &hi.Rating, &hi.Notes, &metadataBytes,
			&hi.IsPublic, &hi.CreatedAt, &hi.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(metadataBytes, &hi.Metadata); err != nil {
			hi.Metadata = map[string]any{}
		}
		items = append(items, hi)
	}
	return items, rows.Err()
}

func (r *postgresHobbyRepo) Save(ctx context.Context, hi *hobby.HobbyItem) error {
	metadataBytes, err := json.Marshal(hi.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal hobby metadata: %w", err)
	}
	query := `
		INSERT INTO hobby_items (id, owner_id, category, title, status, rating, notes, metadata, is_public, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err = r.db.Exec(ctx, query,
		hi.ID, hi.OwnerID, hi.Category, hi.Title, hi.Status, hi.Rating,
		hi.Notes, metadataBytes, hi.IsPublic, hi.CreatedAt, hi.UpdatedAt,
	)
	return err
}

func (r *postgresHobbyRepo) Update(ctx context.Context, hi *hobby.HobbyItem) error {
	metadataBytes, err := json.Marshal(hi.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal hobby metadata for update: %w", err)
	}
	query := `
		UPDATE hobby_items SET
			category = $2, title = $3, status = $4, rating = $5, notes = $6, 
			metadata = $7, is_public = $8, updated_at = NOW()
		WHERE id = $1 AND owner_id = $9
	`
	cmdTag, err := r.db.Exec(ctx, query,
		hi.ID, hi.Category, hi.Title, hi.Status, hi.Rating, hi.Notes,
		metadataBytes, hi.IsPublic, hi.OwnerID,
	)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return hobby.ErrHobbyItemNotFound
	}
	return nil
}

func (r *postgresHobbyRepo) Delete(ctx context.Context, id uuid.UUID, ownerID uuid.UUID) error {
	query := `DELETE FROM hobby_items WHERE id = $1 AND owner_id = $2`
	cmdTag, err := r.db.Exec(ctx, query, id, ownerID)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return hobby.ErrHobbyItemNotFound
	}
	return nil
}

func (r *postgresHobbyRepo) FindByID(ctx context.Context, id uuid.UUID, ownerID uuid.UUID) (*hobby.HobbyItem, error) {
	query := `SELECT * FROM hobby_items WHERE id = $1 AND owner_id = $2`
	row := r.db.QueryRow(ctx, query, id, ownerID)
	return scanHobbyItem(row)
}

func (r *postgresHobbyRepo) ListByOwnerAndCategory(ctx context.Context, ownerID uuid.UUID, category string, limit, offset int) ([]*hobby.HobbyItem, error) {
	builder := psqlHobby.Select("*").
		From("hobby_items").
		Where(sq.Eq{"owner_id": ownerID, "category": category}).
		OrderBy("created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset))

	sql, args, _ := builder.ToSql()
	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query hobby items by owner/category: %w", err)
	}
	return scanHobbyItems(rows)
}

func (r *postgresHobbyRepo) ListPublicByCategory(ctx context.Context, category string, limit, offset int) ([]*hobby.HobbyItem, error) {
	builder := psqlHobby.Select("*").
		From("hobby_items").
		Where(sq.Eq{"is_public": true, "category": category}).
		OrderBy("created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset))

	sql, args, _ := builder.ToSql()
	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query public hobby items by category: %w", err)
	}
	return scanHobbyItems(rows)
}
