package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/khoahotran/personal-os/internal/domain/profile"
)

type postgresProfileRepo struct {
	db *pgxpool.Pool
}

func NewPostgresProfileRepo(db *pgxpool.Pool) profile.Repository {
	return &postgresProfileRepo{db: db}
}

func (r *postgresProfileRepo) GetByUserID(ctx context.Context, ownerID uuid.UUID) (*profile.Profile, error) {
	query := `
		SELECT owner_id, bio, career_timeline, theme_settings, updated_at
		FROM profiles
		WHERE owner_id = $1
	`
	p := &profile.Profile{}
	var careerTimelineBytes, themeSettingsBytes []byte

	err := r.db.QueryRow(ctx, query, ownerID).Scan(
		&p.OwnerID,
		&p.Bio,
		&careerTimelineBytes,
		&themeSettingsBytes,
		&p.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &profile.Profile{
				OwnerID:        ownerID,
				CareerTimeline: []profile.CareerMilestone{},
				ThemeSettings:  map[string]any{},
			}, nil
		}
		return nil, fmt.Errorf("query profile failed: %w", err)
	}

	// Unmarshal JSONB
	if err := json.Unmarshal(careerTimelineBytes, &p.CareerTimeline); err != nil {
		return nil, fmt.Errorf("error when unmarshal career_timeline: %w", err)
	}
	if err := json.Unmarshal(themeSettingsBytes, &p.ThemeSettings); err != nil {
		return nil, fmt.Errorf("error when unmarshal theme_settings: %w", err)
	}

	return p, nil
}

func (r *postgresProfileRepo) Upsert(ctx context.Context, p *profile.Profile) error {
	careerTimelineBytes, err := json.Marshal(p.CareerTimeline)
	if err != nil {
		return fmt.Errorf("error when marshal career_timeline: %w", err)
	}

	themeSettingsBytes, err := json.Marshal(p.ThemeSettings)
	if err != nil {
		return fmt.Errorf("error when marshal theme_settings: %w", err)
	}

	query := `
		INSERT INTO profiles (owner_id, bio, career_timeline, theme_settings, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (owner_id) DO UPDATE SET
			bio = EXCLUDED.bio,
			career_timeline = EXCLUDED.career_timeline,
			updated_at = NOW()
	`
	_, err = r.db.Exec(ctx, query,
		p.OwnerID,
		p.Bio,
		careerTimelineBytes,
		themeSettingsBytes,
		p.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("error when upsert profile: %w", err)
	}
	return nil
}
