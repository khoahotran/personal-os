package persistence

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/khoahotran/personal-os/internal/domain/tag"
)

type postgresTagRepo struct {
	db *pgxpool.Pool
}

func NewPostgresTagRepo(db *pgxpool.Pool) tag.Repository {
	return &postgresTagRepo{db: db}
}

func (r *postgresTagRepo) FindOrCreateTags(ctx context.Context, tagNames []string) ([]tag.Tag, error) {
	if len(tagNames) == 0 {
		return []tag.Tag{}, nil
	}

	tagsToFind := make(map[string]string)
	for _, name := range tagNames {
		slug := strings.ToLower(strings.ReplaceAll(name, " ", "-"))
		tagsToFind[slug] = name
	}

	slugs := make([]string, 0, len(tagsToFind))
	for slug := range tagsToFind {
		slugs = append(slugs, slug)
	}

	insertQuery := `INSERT INTO tags (name, slug) VALUES `
	var inserts []string
	var args []interface{}
	i := 1
	for slug, name := range tagsToFind {
		inserts = append(inserts, fmt.Sprintf("($%d, $%d)", i, i+1))
		args = append(args, name, slug)
		i += 2
	}
	insertQuery += strings.Join(inserts, ",") + " ON CONFLICT (slug) DO NOTHING"

	if _, err := r.db.Exec(ctx, insertQuery, args...); err != nil {
		return nil, fmt.Errorf("failed to bulk insert tags: %w", err)
	}

	query := `SELECT id, name, slug FROM tags WHERE slug = ANY($1)`
	rows, err := r.db.Query(ctx, query, slugs)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve tags: %w", err)
	}
	defer rows.Close()

	tags := make([]tag.Tag, 0)
	for rows.Next() {
		var t tag.Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}

func (r *postgresTagRepo) SetTagsForResource(ctx context.Context, resourceID uuid.UUID, resourceType string, tagIDs []uuid.UUID) error {

	deleteQuery := `DELETE FROM tag_relations WHERE resource_id = $1 AND resource_type = $2`
	if _, err := r.db.Exec(ctx, deleteQuery, resourceID, resourceType); err != nil {
		return fmt.Errorf("failed to delete old tags: %w", err)
	}

	if len(tagIDs) == 0 {
		return nil
	}

	rowsToInsert := make([][]interface{}, len(tagIDs))
	for i, tagID := range tagIDs {
		rowsToInsert[i] = []interface{}{tagID, resourceID, resourceType}
	}

	_, err := r.db.CopyFrom(
		ctx,
		pgx.Identifier{"tag_relations"},
		[]string{"tag_id", "resource_id", "resource_type"},
		pgx.CopyFromRows(rowsToInsert),
	)

	if err != nil {
		return fmt.Errorf("failed to set new tags: %w", err)
	}
	return nil
}

func (r *postgresTagRepo) GetTagsForResource(ctx context.Context, resourceID uuid.UUID, resourceType string) ([]tag.Tag, error) {
	query := `
		SELECT t.id, t.name, t.slug 
		FROM tags t
		JOIN tag_relations tr ON t.id = tr.tag_id
		WHERE tr.resource_id = $1 AND tr.resource_type = $2
	`
	rows, err := r.db.Query(ctx, query, resourceID, resourceType)
	if err != nil {
		return nil, fmt.Errorf("failed to get tags for resource: %w", err)
	}
	defer rows.Close()

	tags := make([]tag.Tag, 0)
	for rows.Next() {
		var t tag.Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}
