package profile

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/khoahotran/personal-os/internal/domain/profile"
)

type ProfileUseCase struct {
	profileRepo profile.Repository
}

func NewProfileUseCase(repo profile.Repository) *ProfileUseCase {
	return &ProfileUseCase{
		profileRepo: repo,
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
		return nil, fmt.Errorf("get profile failed: %w", err)
	}
	return &GetProfileOutput{Profile: p}, nil
}

type UpdateProfileInput struct {
	OwnerID        uuid.UUID
	Bio            string
	CareerTimeline []profile.CareerMilestone
}

type UpdateProfileOutput struct {
	Profile *profile.Profile
}

func (uc *ProfileUseCase) ExecuteUpdateProfile(ctx context.Context, input UpdateProfileInput) (*UpdateProfileOutput, error) {
	p := &profile.Profile{
		OwnerID:        input.OwnerID,
		Bio:            input.Bio,
		CareerTimeline: input.CareerTimeline,
		UpdatedAt:      time.Now().UTC(),
	}

	err := uc.profileRepo.Upsert(ctx, p)
	if err != nil {
		return nil, fmt.Errorf("update profile failed: %w", err)
	}

	return &UpdateProfileOutput{Profile: p}, nil
}
