package http

import (
	"time"
	"github.com/khoahotran/personal-os/internal/domain/profile"
)

// --- Profile DTOs ---

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

// --- Helper Converters ---

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