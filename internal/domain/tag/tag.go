package tag

import (
	"context"

	"github.com/google/uuid"
)

type Tag struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Slug string    `json:"slug"`
}

type TagRelation struct {
	TagID        uuid.UUID `json:"tag_id"`
	ResourceID   uuid.UUID `json:"resource_id"`
	ResourceType string    `json:"resource_type"`
}

type Repository interface {
	FindOrCreateTags(ctx context.Context, tagNames []string) ([]Tag, error)
	SetTagsForResource(ctx context.Context, resourceID uuid.UUID, resourceType string, tagIDs []uuid.UUID) error
	GetTagsForResource(ctx context.Context, resourceID uuid.UUID, resourceType string) ([]Tag, error)
}
