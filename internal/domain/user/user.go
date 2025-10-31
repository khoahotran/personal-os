package user

import (
	"context"

	"github.com/google/uuid"
)

type User struct {
	ID              uuid.UUID      `json:"id"`
	Email           string         `json:"email"`
	Name            *string        `json:"name"`
	PasswordHash    string         `json:"-"`
	ProfileSettings map[string]any `json:"profile_settings"`
}

type Repository interface {
	FindByEmail(ctx context.Context, email string) (*User, error)
}
