package persistence

import (
	"context"
	"encoding/json"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/khoahotran/personal-os/internal/domain/hobby"
	"github.com/khoahotran/personal-os/pkg/apperror"
	"github.com/khoahotran/personal-os/pkg/logger"
	"go.uber.org/zap"
)

type postgresHobbyRepo struct {
	db     *pgxpool.Pool
	logger logger.Logger
}

func NewPostgresHobbyRepo(db *pgxpool.Pool, logger logger.Logger) hobby.Repository {
	return &postgresHobbyRepo{db: db, logger: logger}
}

var psqlHobby = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

func scanHobbyItem(row pgx.Row, l logger.Logger) (*hobby.HobbyItem, error) {
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
		return nil, apperror.NewInternal("failed to scan hobby item row", err)
	}

	if err := json.Unmarshal(metadataBytes, &hi.Metadata); err != nil {
		l.Warn("Failed to unmarshal hobby metadata", zap.String("item_id", hi.ID.String()), zap.Error(err))
		hi.Metadata = map[string]any{}
	}
	return hi, nil
}

func scanHobbyItems(rows pgx.Rows, l logger.Logger) ([]*hobby.HobbyItem, error) {
	defer rows.Close()
	items := make([]*hobby.HobbyItem, 0)
	for rows.Next() {

		hi, err := scanHobbyItem(rows, l)
		if err != nil {
			return nil, err
		}
		items = append(items, hi)
	}
	if err := rows.Err(); err != nil {
		return nil, apperror.NewInternal("error iterating hobby rows", err)
	}
	return items, nil
}

func (r *postgresHobbyRepo) Save(ctx context.Context, hi *hobby.HobbyItem) error {
	metadataBytes, err := json.Marshal(hi.Metadata)
	if err != nil {
		return apperror.NewInternal("failed to marshal hobby metadata", err)
	}
	query := `
		INSERT INTO hobby_items (id, owner_id, category, title, status, rating, notes, metadata, is_public, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err = r.db.Exec(ctx, query,
		hi.ID, hi.OwnerID, hi.Category, hi.Title, hi.Status, hi.Rating,
		hi.Notes, metadataBytes, hi.IsPublic, hi.CreatedAt, hi.UpdatedAt,
	)
	if err != nil {
		return apperror.NewInternal("failed to save hobby item", err)
	}
	return err
}

func (r *postgresHobbyRepo) Update(ctx context.Context, hi *hobby.HobbyItem) error {
	metadataBytes, err := json.Marshal(hi.Metadata)
	if err != nil {
		return apperror.NewInternal("failed to marshal hobby metadata", err)
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
		return apperror.NewInternal("failed to update hobby item", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return apperror.NewNotFound("hobby item", hi.ID.String())
	}
	return nil
}

func (r *postgresHobbyRepo) Delete(ctx context.Context, id uuid.UUID, ownerID uuid.UUID) error {
	query := `DELETE FROM hobby_items WHERE id = $1 AND owner_id = $2`
	cmdTag, err := r.db.Exec(ctx, query, id, ownerID)
	if err != nil {
		return apperror.NewInternal("failed to delete hobby item", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return apperror.NewNotFound("hobby item", id.String())
	}
	return nil
}

func (r *postgresHobbyRepo) FindByID(ctx context.Context, id uuid.UUID, ownerID uuid.UUID) (*hobby.HobbyItem, error) {
	query := `SELECT * FROM hobby_items WHERE id = $1 AND owner_id = $2`
	row := r.db.QueryRow(ctx, query, id, ownerID)
	return scanHobbyItem(row, r.logger)
}

func (r *postgresHobbyRepo) ListByOwnerAndCategory(ctx context.Context, ownerID uuid.UUID, category string, limit, offset int) ([]*hobby.HobbyItem, error) {
	builder := psqlHobby.Select("*").
		From("hobby_items").
		Where(sq.Eq{"owner_id": ownerID, "category": category}).
		OrderBy("created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset))

	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, apperror.NewInternal("failed to build list hobby by owner query", err)
	}
	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, apperror.NewInternal("failed to query hobby items by owner/category", err)
	}
	return scanHobbyItems(rows, r.logger)
}

func (r *postgresHobbyRepo) ListPublicByCategory(ctx context.Context, category string, limit, offset int) ([]*hobby.HobbyItem, error) {
	builder := psqlHobby.Select("*").
		From("hobby_items").
		Where(sq.Eq{"is_public": true, "category": category}).
		OrderBy("created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset))

	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, apperror.NewInternal("failed to build list public hobby items query", err)
	}
	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, apperror.NewInternal("failed to query public hobby items by category", err)
	}
	return scanHobbyItems(rows, r.logger)
}
