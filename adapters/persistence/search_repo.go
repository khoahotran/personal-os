package persistence

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/khoahotran/personal-os/internal/domain/search"
	"github.com/khoahotran/personal-os/pkg/apperror"
	"github.com/khoahotran/personal-os/pkg/logger"
)

type postgresSearchRepo struct {
	db     *pgxpool.Pool
	logger logger.Logger
}

func NewPostgresSearchRepo(db *pgxpool.Pool, logger logger.Logger) search.Repository {
	return &postgresSearchRepo{db: db, logger: logger}
}

var psqlSearch = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

func (r *postgresSearchRepo) searchBase(ctx context.Context, builder sq.SelectBuilder) ([]search.SearchResult, error) {
	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, apperror.NewInternal("failed to build search query", err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, apperror.NewInternal("failed to execute search query", err)
	}
	defer rows.Close()

	results := make([]search.SearchResult, 0)
	for rows.Next() {
		var res search.SearchResult
		if err := rows.Scan(
			&res.ID, &res.ResourceType, &res.Title, &res.Slug,
			&res.Snippet, &res.Rank, &res.IsPublic, &res.UpdatedAt,
		); err != nil {
			return nil, apperror.NewInternal("failed to scan search result", err)
		}
		results = append(results, res)
	}
	if err := rows.Err(); err != nil {
		return nil, apperror.NewInternal("error iterating search results", err)
	}
	return results, nil
}

func (r *postgresSearchRepo) SearchPrivate(ctx context.Context, query string, ownerID uuid.UUID, limit int) ([]search.SearchResult, error) {
	finalSql := `
	(SELECT 
		id, 'post' AS resource_type, title, slug,
		ts_headline('simple', content_markdown, to_tsquery('simple', $1), 'StartSel=*,StopSel=*,MaxFragments=1,MaxWords=10,MinWords=5') AS snippet,
		ts_rank_cd(ts, to_tsquery('simple', $1)) AS rank,
		(status = 'public') AS is_public,
		updated_at
	FROM posts
	WHERE owner_id = $2 AND ts @@ to_tsquery('simple', $1))
	
	UNION ALL
	
	(SELECT 
		id, 'project' AS resource_type, title, slug,
		ts_headline('simple', description, to_tsquery('simple', $1), 'StartSel=*,StopSel=*,MaxFragments=1,MaxWords=10,MinWords=5') AS snippet,
		ts_rank_cd(ts, to_tsquery('simple', $1)) AS rank,
		is_public,
		updated_at
	FROM projects
	WHERE owner_id = $2 AND ts @@ to_tsquery('simple', $1))

	-- (Sau này có thể UNION ALL với bảng hobby_items ở đây)
	
	ORDER BY rank DESC
	LIMIT $3
	`

	finalArgs := []interface{}{query, ownerID, limit}

	rows, err := r.db.Query(ctx, finalSql, finalArgs...)
	if err != nil {
		return nil, apperror.NewInternal("failed to execute private search", err)
	}
	defer rows.Close()

	results := make([]search.SearchResult, 0)
	for rows.Next() {
		var res search.SearchResult
		if err := rows.Scan(
			&res.ID, &res.ResourceType, &res.Title, &res.Slug,
			&res.Snippet, &res.Rank, &res.IsPublic, &res.UpdatedAt,
		); err != nil {
			return nil, apperror.NewInternal("failed to scan search result", err)
		}
		results = append(results, res)
	}
	if err := rows.Err(); err != nil {
		return nil, apperror.NewInternal("error iterating search results", err)
	}
	return results, nil
}

func (r *postgresSearchRepo) SearchPublic(ctx context.Context, query string, limit int) ([]search.SearchResult, error) {
	finalSql := `
	(SELECT 
		id, 'post' AS resource_type, title, slug,
		ts_headline('simple', content_markdown, to_tsquery('simple', $1), 'StartSel=*,StopSel=*,MaxFragments=1,MaxWords=10,MinWords=5') AS snippet,
		ts_rank_cd(ts, to_tsquery('simple', $1)) AS rank,
		true AS is_public,
		updated_at
	FROM posts
	WHERE status = 'public' AND ts @@ to_tsquery('simple', $1))
	
	UNION ALL
	
	(SELECT 
		id, 'project' AS resource_type, title, slug,
		ts_headline('simple', description, to_tsquery('simple', $1), 'StartSel=*,StopSel=*,MaxFragments=1,MaxWords=10,MinWords=5') AS snippet,
		ts_rank_cd(ts, to_tsquery('simple', $1)) AS rank,
		is_public,
		updated_at
	FROM projects
	WHERE is_public = true AND ts @@ to_tsquery('simple', $1))
	
	ORDER BY rank DESC
	LIMIT $2
	`
	finalArgs := []interface{}{query, limit}

	rows, err := r.db.Query(ctx, finalSql, finalArgs...)
	if err != nil {
		return nil, apperror.NewInternal("failed to execute public search", err)
	}
	defer rows.Close()

	results := make([]search.SearchResult, 0)
	for rows.Next() {
		var res search.SearchResult
		if err := rows.Scan(
			&res.ID, &res.ResourceType, &res.Title, &res.Slug,
			&res.Snippet, &res.Rank, &res.IsPublic, &res.UpdatedAt,
		); err != nil {
			return nil, apperror.NewInternal("failed to scan search result", err)
		}
		results = append(results, res)
	}
	if err := rows.Err(); err != nil {
		return nil, apperror.NewInternal("error iterating search results", err)
	}
	return results, nil
}
