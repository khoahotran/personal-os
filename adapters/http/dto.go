package http

import (
	"time"

	"github.com/khoahotran/personal-os/internal/domain/post"
	"github.com/khoahotran/personal-os/internal/domain/profile"
	"github.com/khoahotran/personal-os/internal/domain/tag"
)

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

func (r *CreatePostRequest) ToDomainPostStatus() post.PostStatus {
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
