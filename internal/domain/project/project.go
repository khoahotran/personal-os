package project

import (
	"context"
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
)

type Project struct {
	ID            uuid.UUID      `json:"id"`
	OwnerID       uuid.UUID      `json:"owner_id"`
	Slug          string         `json:"slug"`
	Title         string         `json:"title"`
	Description   string         `json:"description"`
	Stack         []string       `json:"stack"`
	RepositoryURL *string        `json:"repository_url"`
	LiveURL       *string        `json:"live_url"`
	Media         []ProjectMedia `json:"media"`
	IsPublic      bool           `json:"is_public"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

type ProjectMedia struct {
	Type string `json:"type"`
	URL  string `json:"url"`
	Alt  string `json:"alt"`
}

var (
	ErrInvalidSlug     = errors.New("slug only allows lowercase letters, numbers, and hyphens")
	ErrProjectNotFound = errors.New("project not found")
	slugRegex          = regexp.MustCompile(`^[a-z0-9-]+$`)
)

func (p *Project) Validate() error {
	if !slugRegex.MatchString(p.Slug) {
		return ErrInvalidSlug
	}
	return nil
}

type Repository interface {
	Save(ctx context.Context, project *Project) error
	Update(ctx context.Context, project *Project) error
	Delete(ctx context.Context, id uuid.UUID, ownerID uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID, ownerID uuid.UUID) (*Project, error)
	FindBySlug(ctx context.Context, slug string) (*Project, error)
	FindPublicBySlug(ctx context.Context, slug string) (*Project, error)
	ListByOwner(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]*Project, error)
	ListPublic(ctx context.Context, limit, offset int) ([]*Project, error)
}
