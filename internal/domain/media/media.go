package media

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type MediaStatus string

const (
	StatusPending MediaStatus = "pending"
	StatusReady   MediaStatus = "ready"
	StatusError   MediaStatus = "error"
)

type Media struct {
	ID           uuid.UUID      `json:"id"`
	OwnerID      uuid.UUID      `json:"owner_id"`
	Provider     string         `json:"provider"`
	URL          string         `json:"url"`
	ThumbnailURL *string        `json:"thumbnail_url"`
	Status       MediaStatus    `json:"status"`
	Metadata     map[string]any `json:"metadata"`
	IsPublic     bool           `json:"is_public"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

type Repository interface {
	Save(ctx context.Context, media *Media) error
	Update(ctx context.Context, media *Media) error
	Delete(ctx context.Context, id uuid.UUID, ownerID uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID, ownerID uuid.UUID) (*Media, error)
	ListPublic(ctx context.Context, limit, offset int) ([]*Media, error)
	ListByOwner(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]*Media, error)
}
