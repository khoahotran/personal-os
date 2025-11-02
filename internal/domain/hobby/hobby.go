package hobby

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

const (
	CategoryGame  = "game"
	CategoryMovie = "movie"
	CategoryBook  = "book"
	CategoryAnime = "anime"
)

type HobbyItem struct {
	ID        uuid.UUID      `json:"id"`
	OwnerID   uuid.UUID      `json:"owner_id"`
	Category  string         `json:"category"`
	Title     string         `json:"title"`
	Status    string         `json:"status"`
	Rating    int            `json:"rating"`
	Notes     string         `json:"notes"`
	Metadata  map[string]any `json:"metadata"`
	IsPublic  bool           `json:"is_public"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

var (
	ErrHobbyItemNotFound = errors.New("hobby item not found")
	ErrInvalidCategory   = errors.New("invalid hobby category")
)

func (hi *HobbyItem) Validate() error {
	if hi.Title == "" {
		return errors.New("title is required")
	}
	switch hi.Category {
	case CategoryGame, CategoryMovie, CategoryBook, CategoryAnime:

	default:
		return ErrInvalidCategory
	}
	return nil
}

type Repository interface {
	Save(ctx context.Context, item *HobbyItem) error
	Update(ctx context.Context, item *HobbyItem) error
	Delete(ctx context.Context, id uuid.UUID, ownerID uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID, ownerID uuid.UUID) (*HobbyItem, error)
	ListByOwnerAndCategory(ctx context.Context, ownerID uuid.UUID, category string, limit, offset int) ([]*HobbyItem, error)
	ListPublicByCategory(ctx context.Context, category string, limit, offset int) ([]*HobbyItem, error)
}
