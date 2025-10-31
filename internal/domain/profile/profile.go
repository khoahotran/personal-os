package profile

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type CareerMilestone struct {
	Date        time.Time `json:"date"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
}

type Profile struct {
	OwnerID        uuid.UUID         `json:"owner_id"`
	Bio            string            `json:"bio"`
	CareerTimeline []CareerMilestone `json:"career_timeline"`
	ThemeSettings  map[string]any    `json:"theme_settings"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

type Repository interface {
	GetByUserID(ctx context.Context, ownerID uuid.UUID) (*Profile, error)
	Upsert(ctx context.Context, profile *Profile) error
}
