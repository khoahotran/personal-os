package http

import (
	"time"

	"github.com/khoahotran/personal-os/internal/domain/hobby"
	"github.com/khoahotran/personal-os/internal/domain/media"
	"github.com/khoahotran/personal-os/internal/domain/post"
	"github.com/khoahotran/personal-os/internal/domain/profile"
	"github.com/khoahotran/personal-os/internal/domain/project"
	"github.com/khoahotran/personal-os/internal/domain/search"
	"github.com/khoahotran/personal-os/internal/domain/tag"
)

// Profile DTOs
type CareerMilestoneDTO struct {
	Date        time.Time `json:"date"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
}

type ProfileDTO struct {
	Bio            string               `json:"bio"`
	CareerTimeline []CareerMilestoneDTO `json:"career_timeline"`
	ThemeSettings  map[string]any       `json:"theme_settings"`
	UpdatedAt      time.Time            `json:"updated_at"`
}

type UpdateProfileRequest struct {
	Bio            string `json:"bio"`
	CareerTimeline []struct {
		Date        time.Time `json:"date" binding:"required"`
		Title       string    `json:"title" binding:"required"`
		Description string    `json:"description"`
	} `json:"career_timeline"`
	ThemeSettings map[string]any `json:"theme_settings"`
}

func ToProfileDTO(p *profile.Profile) ProfileDTO {
	dto := ProfileDTO{
		Bio:           p.Bio,
		ThemeSettings: p.ThemeSettings,
		UpdatedAt:     p.UpdatedAt,
	}
	dto.CareerTimeline = make([]CareerMilestoneDTO, len(p.CareerTimeline))
	for i, m := range p.CareerTimeline {
		dto.CareerTimeline[i] = CareerMilestoneDTO{
			Date:        m.Date,
			Title:       m.Title,
			Description: m.Description,
		}
	}
	return dto
}

func (req *UpdateProfileRequest) ToDomainMilestones() []profile.CareerMilestone {
	domainMilestones := make([]profile.CareerMilestone, len(req.CareerTimeline))
	for i, m := range req.CareerTimeline {
		domainMilestones[i] = profile.CareerMilestone{
			Date:        m.Date,
			Title:       m.Title,
			Description: m.Description,
		}
	}
	return domainMilestones
}

// Post DTOs

type CreatePostRequest struct {
	Title   string   `json:"title" binding:"required"`
	Content string   `json:"content"`
	Slug    string   `json:"slug"`
	Status  string   `json:"status" binding:"required,oneof=draft private public"`
	Tags    []string `json:"tags"`
}

type PostDTO struct {
	ID              string     `json:"id"`
	Slug            string     `json:"slug"`
	Title           string     `json:"title"`
	ContentMarkdown string     `json:"content_markdown"`
	Status          string     `json:"status"`
	OgImageURL      *string    `json:"og_image_url"`
	PublishedAt     *time.Time `json:"published_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	Tags            []string   `json:"tags"`
}

type UpdatePostRequest struct {
	Title   string   `json:"title" binding:"required"`
	Content string   `json:"content"`
	Slug    string   `json:"slug" binding:"required"`
	Status  string   `json:"status" binding:"required,oneof=draft private public"`
	Tags    []string `json:"tags"`
}

func (r *UpdatePostRequest) ToDomainPostStatus() post.PostStatus {
	switch r.Status {
	case "public":
		return post.StatusPublic
	case "private":
		return post.StatusPrivate
	default:
		return post.StatusDraft
	}
}

type PostSummaryDTO struct {
	ID          string     `json:"id"`
	Slug        string     `json:"slug"`
	Title       string     `json:"title"`
	Status      string     `json:"status"`
	OgImageURL  *string    `json:"og_image_url,omitempty"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func ToPostSummaryDTO(p *post.Post) PostSummaryDTO {
	return PostSummaryDTO{
		ID:          p.ID.String(),
		Slug:        p.Slug,
		Title:       p.Title,
		Status:      string(p.Status),
		OgImageURL:  p.OgImageURL,
		PublishedAt: p.PublishedAt,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func ToPostDTO(p *post.Post, tags []tag.Tag) PostDTO {
	tagNames := make([]string, len(tags))
	for i, t := range tags {
		tagNames[i] = t.Name
	}

	return PostDTO{
		ID:              p.ID.String(),
		Slug:            p.Slug,
		Title:           p.Title,
		ContentMarkdown: p.ContentMarkdown,
		Status:          string(p.Status),
		OgImageURL:      p.OgImageURL,
		PublishedAt:     p.PublishedAt,
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
		Tags:            tagNames,
	}
}

// Project DTOs

type CreateProjectRequest struct {
	Title         string   `json:"title" binding:"required"`
	Slug          string   `json:"slug"`
	Description   string   `json:"description"`
	Stack         []string `json:"stack"`
	RepositoryURL *string  `json:"repository_url"`
	LiveURL       *string  `json:"live_url"`
	IsPublic      bool     `json:"is_public"`
	TagNames      []string `json:"tags"`
}

type ProjectMediaDTO struct {
	Type string `json:"type"`
	URL  string `json:"url"`
	Alt  string `json:"alt"`
}

type ProjectSummaryDTO struct {
	ID        string    `json:"id"`
	Slug      string    `json:"slug"`
	Title     string    `json:"title"`
	Stack     []string  `json:"stack"`
	IsPublic  bool      `json:"is_public"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ProjectDTO struct {
	ID            string            `json:"id"`
	Slug          string            `json:"slug"`
	Title         string            `json:"title"`
	Description   string            `json:"description"`
	Stack         []string          `json:"stack"`
	RepositoryURL *string           `json:"repository_url"`
	LiveURL       *string           `json:"live_url"`
	Media         []ProjectMediaDTO `json:"media"`
	IsPublic      bool              `json:"is_public"`
	UpdatedAt     time.Time         `json:"updated_at"`
	Tags          []string          `json:"tags"`
}

type UpdateProjectRequest struct {
	Title         string   `json:"title" binding:"required"`
	Slug          string   `json:"slug" binding:"required"`
	Description   string   `json:"description"`
	Stack         []string `json:"stack"`
	RepositoryURL *string  `json:"repository_url"`
	LiveURL       *string  `json:"live_url"`
	IsPublic      bool     `json:"is_public"`
	TagNames      []string `json:"tags"`
}

func ToProjectDTO(p *project.Project, tags []tag.Tag) ProjectDTO {
	tagNames := make([]string, len(tags))
	for i, t := range tags {
		tagNames[i] = t.Name
	}
	mediaDTOs := make([]ProjectMediaDTO, len(p.Media))
	for i, m := range p.Media {
		mediaDTOs[i] = ProjectMediaDTO(m)
	}

	return ProjectDTO{
		ID:            p.ID.String(),
		Slug:          p.Slug,
		Title:         p.Title,
		Description:   p.Description,
		Stack:         p.Stack,
		RepositoryURL: p.RepositoryURL,
		LiveURL:       p.LiveURL,
		Media:         mediaDTOs,
		IsPublic:      p.IsPublic,
		UpdatedAt:     p.UpdatedAt,
		Tags:          tagNames,
	}
}

func ToProjectSummaryDTO(p *project.Project) ProjectSummaryDTO {
	return ProjectSummaryDTO{
		ID:        p.ID.String(),
		Slug:      p.Slug,
		Title:     p.Title,
		Stack:     p.Stack,
		IsPublic:  p.IsPublic,
		UpdatedAt: p.UpdatedAt,
	}
}

// Media DTOs
type MediaDTO struct {
	ID           string         `json:"id"`
	URL          string         `json:"url"`
	ThumbnailURL *string        `json:"thumbnail_url,omitempty"`
	Status       string         `json:"status"`
	Metadata     map[string]any `json:"metadata"`
	IsPublic     bool           `json:"is_public"`
	CreatedAt    time.Time      `json:"created_at"`
}

type UpdateMediaRequest struct {
	Metadata map[string]any `json:"metadata"`
	IsPublic bool           `json:"is_public"`
}

func ToMediaDTO(m *media.Media) MediaDTO {
	return MediaDTO{
		ID:           m.ID.String(),
		URL:          m.URL,
		ThumbnailURL: m.ThumbnailURL,
		Status:       string(m.Status),
		Metadata:     m.Metadata,
		IsPublic:     m.IsPublic,
		CreatedAt:    m.CreatedAt,
	}
}

// Hobby DTOs

type HobbyItemDTO struct {
	ID        string         `json:"id"`
	Category  string         `json:"category"`
	Title     string         `json:"title"`
	Status    string         `json:"status"`
	Rating    int            `json:"rating"`
	Notes     string         `json:"notes"`
	Metadata  map[string]any `json:"metadata"`
	IsPublic  bool           `json:"is_public"`
	UpdatedAt time.Time      `json:"updated_at"`
}

type CreateOrUpdateHobbyItemRequest struct {
	Category string         `json:"category" binding:"required"`
	Title    string         `json:"title" binding:"required"`
	Status   string         `json:"status"`
	Rating   int            `json:"rating" binding:"min=0,max=10"`
	Notes    string         `json:"notes"`
	Metadata map[string]any `json:"metadata"`
	IsPublic bool           `json:"is_public"`
}

func ToHobbyItemDTO(hi *hobby.HobbyItem) HobbyItemDTO {
	return HobbyItemDTO{
		ID:        hi.ID.String(),
		Category:  hi.Category,
		Title:     hi.Title,
		Status:    hi.Status,
		Rating:    hi.Rating,
		Notes:     hi.Notes,
		Metadata:  hi.Metadata,
		IsPublic:  hi.IsPublic,
		UpdatedAt: hi.UpdatedAt,
	}
}

type ChatRequest struct {
	Query string `json:"query" binding:"required"`
	Limit int    `json:"limit"`
}

type ChatResponse struct {
	Response string           `json:"response"`
	Sources  []PostSummaryDTO `json:"sources"`
}

type SearchResultDTO struct {
	ID           string    `json:"id"`
	ResourceType string    `json:"resource_type"`
	Title        string    `json:"title"`
	Slug         string    `json:"slug"`
	Snippet      string    `json:"snippet"`
	Rank         float32   `json:"rank"`
	IsPublic     bool      `json:"is_public"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func ToSearchResultDTO(s search.SearchResult) SearchResultDTO {
	return SearchResultDTO{
		ID:           s.ID.String(),
		ResourceType: s.ResourceType,
		Title:        s.Title,
		Slug:         s.Slug,
		Snippet:      s.Snippet,
		Rank:         s.Rank,
		IsPublic:     s.IsPublic,
		UpdatedAt:    s.UpdatedAt,
	}
}
