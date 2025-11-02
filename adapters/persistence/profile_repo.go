package persistence

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/khoahotran/personal-os/internal/domain/profile"
	"github.com/khoahotran/personal-os/pkg/apperror"
	"github.com/khoahotran/personal-os/pkg/logger"
)

type postgresProfileRepo struct {
	db     *pgxpool.Pool
	logger logger.Logger
}

func NewPostgresProfileRepo(db *pgxpool.Pool, logger logger.Logger) profile.Repository {
	return &postgresProfileRepo{db: db, logger: logger}
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
		return nil, apperror.NewInternal("failed to query profile", err)
	}

	// Unmarshal JSONB
	if err := json.Unmarshal(careerTimelineBytes, &p.CareerTimeline); err != nil {
		r.logger.Warn("Failed to unmarshal career_timeline", zap.String("owner_id", ownerID.String()), zap.Error(err))
		p.CareerTimeline = []profile.CareerMilestone{}
	}
	if err := json.Unmarshal(themeSettingsBytes, &p.ThemeSettings); err != nil {
		r.logger.Warn("Failed to unmarshal theme_settings", zap.String("owner_id", ownerID.String()), zap.Error(err))
		p.ThemeSettings = map[string]any{}
	}

	return p, nil
}

func (r *postgresProfileRepo) Upsert(ctx context.Context, p *profile.Profile) error {
	careerTimelineBytes, err := json.Marshal(p.CareerTimeline)
	if err != nil {
		return apperror.NewInternal("failed to marshal career_timeline", err)
	}

	themeSettingsBytes, err := json.Marshal(p.ThemeSettings)
	if err != nil {
		return apperror.NewInternal("failed to marshal theme_settings", err)
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
		return apperror.NewInternal("failed to upsert profile", err)
	}
	return nil
}
