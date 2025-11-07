package search

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type SearchResult struct {
	ID           uuid.UUID `json:"id"`
	ResourceType string    `json:"resource_type"`
	Title        string    `json:"title"`
	Slug         string    `json:"slug"`
	Snippet      string    `json:"snippet"`
	Rank         float32   `json:"rank"`
	IsPublic     bool      `json:"is_public"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Repository interface {
	SearchPublic(ctx context.Context, query string, limit int) ([]SearchResult, error)

	SearchPrivate(ctx context.Context, query string, ownerID uuid.UUID, limit int) ([]SearchResult, error)
}
