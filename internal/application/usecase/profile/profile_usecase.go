package profile

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/khoahotran/personal-os/internal/domain/profile"
	"github.com/khoahotran/personal-os/pkg/apperror"
	"github.com/khoahotran/personal-os/pkg/logger"
	"go.uber.org/zap"
)

type ProfileUseCase struct {
	profileRepo profile.Repository
	logger      logger.Logger
}

func NewProfileUseCase(repo profile.Repository, log logger.Logger) *ProfileUseCase {
	return &ProfileUseCase{
		profileRepo: repo,
		logger:      log,
	}
}

type GetProfileInput struct {
	OwnerID uuid.UUID
}

type GetProfileOutput struct {
	Profile *profile.Profile
}

func (uc *ProfileUseCase) ExecuteGetProfile(ctx context.Context, input GetProfileInput) (*GetProfileOutput, error) {
	p, err := uc.profileRepo.GetByUserID(ctx, input.OwnerID)
	if err != nil {
		uc.logger.Error("Failed to get profile", err, zap.String("owner_id", input.OwnerID.String()))
		return nil, err
	}
	return &GetProfileOutput{Profile: p}, nil
}

type UpdateProfileInput struct {
	OwnerID        uuid.UUID
	Bio            string
	CareerTimeline []profile.CareerMilestone
	ThemeSettings  map[string]any
}

type UpdateProfileOutput struct {
	Profile *profile.Profile
}

func (uc *ProfileUseCase) ExecuteUpdateProfile(ctx context.Context, input UpdateProfileInput) (*UpdateProfileOutput, error) {
	p, err := uc.profileRepo.GetByUserID(ctx, input.OwnerID)
	if err != nil {
		return nil, err
	}

	p.Bio = input.Bio
	p.CareerTimeline = input.CareerTimeline
	p.UpdatedAt = time.Now().UTC()

	if input.ThemeSettings != nil {
		p.ThemeSettings = input.ThemeSettings
	}

	if err := uc.profileRepo.Upsert(ctx, p); err != nil {
		uc.logger.Error("Failed to update profile", err, zap.String("owner_id", input.OwnerID.String()))
		return nil, apperror.NewInternal("failed to update profile", err)
	}

	return &UpdateProfileOutput{Profile: p}, nil
}
