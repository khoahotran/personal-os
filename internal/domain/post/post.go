package post

import (
	"context"
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
)

type PostStatus string

const (
	StatusDraft   PostStatus = "draft"
	StatusPrivate PostStatus = "private"
	StatusPublic  PostStatus = "public"
	StatusPending PostStatus = "pending"
)

type PostVersion struct {
	ID          uuid.UUID `json:"id"`
	PostID      uuid.UUID `json:"post_id"`
	ContentDiff string    `json:"content_diff"`
	CreatedAt   time.Time `json:"created_at"`
}
type Post struct {
	ID              uuid.UUID      `json:"id"`
	OwnerID         uuid.UUID      `json:"owner_id"`
	Slug            string         `json:"slug"`
	Title           string         `json:"title"`
	ContentMarkdown string         `json:"content_markdown"`
	Status          PostStatus     `json:"status"`
	OgImageURL      *string        `json:"og_image_url"`
	ThumbnailURL    *string        `json:"thumbnail_url"`
	VersionHistory  []PostVersion  `json:"version_history"`
	Metadata        map[string]any `json:"metadata"`
	PublishedAt     *time.Time     `json:"published_at"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

var (
	ErrInvalidPostStatus = errors.New("invalid status")
	ErrInvalidPostSlug   = errors.New("slug only includes lowercase letter, digit and -")
	postSlugRegex        = regexp.MustCompile(`^[a-z0-9-]+$`)
	ErrPostNotFound      = errors.New("post not found")
)

func (p *Post) Validate() error {
	if !postSlugRegex.MatchString(p.Slug) {
		return ErrInvalidPostSlug
	}
	switch p.Status {
	case StatusDraft, StatusPrivate, StatusPublic, StatusPending:
	default:
		return ErrInvalidPostStatus
	}
	return nil
}

func (p *Post) AddVersion(timestamp time.Time, oldContent string) {
	const maxHistory = 10

	newVersion := PostVersion{
		ID:          uuid.New(),
		PostID:      p.ID,
		ContentDiff: oldContent,
		CreatedAt:   timestamp,
	}

	p.VersionHistory = append([]PostVersion{newVersion}, p.VersionHistory...)

	if len(p.VersionHistory) > maxHistory {
		p.VersionHistory = p.VersionHistory[:maxHistory]
	}
}

func (p *Post) MarkAsReady(imageURL, thumbnailURL string) {
	p.OgImageURL = &imageURL
	p.ThumbnailURL = &thumbnailURL
}

type Repository interface {
	Save(ctx context.Context, post *Post) error
	Update(ctx context.Context, post *Post) error
	Delete(ctx context.Context, id uuid.UUID, ownerID uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID, ownerID uuid.UUID) (*Post, error)
	FindBySlug(ctx context.Context, slug string) (*Post, error)
	FindPublicBySlug(ctx context.Context, slug string) (*Post, error)
	ListByOwner(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]*Post, error)
	ListPublic(ctx context.Context, limit, offset int) ([]*Post, error)
}
