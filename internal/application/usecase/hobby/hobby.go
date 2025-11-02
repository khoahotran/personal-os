package hobby

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/khoahotran/personal-os/internal/domain/hobby"
	"github.com/khoahotran/personal-os/pkg/apperror"
	"github.com/khoahotran/personal-os/pkg/logger"
)

type HobbyUseCase struct {
	repo   hobby.Repository
	logger logger.Logger
}

func NewHobbyUseCase(r hobby.Repository, log logger.Logger) *HobbyUseCase {
	return &HobbyUseCase{repo: r, logger: log}
}

type CreateHobbyItemInput struct {
	OwnerID  uuid.UUID
	Category string
	Title    string
	Status   string
	Rating   int
	Notes    string
	Metadata map[string]any
	IsPublic bool
}

func (uc *HobbyUseCase) CreateHobbyItem(ctx context.Context, in CreateHobbyItemInput) (*hobby.HobbyItem, error) {
	now := time.Now().UTC()
	item := &hobby.HobbyItem{
		ID:        uuid.New(),
		OwnerID:   in.OwnerID,
		Category:  in.Category,
		Title:     in.Title,
		Status:    in.Status,
		Rating:    in.Rating,
		Notes:     in.Notes,
		Metadata:  in.Metadata,
		IsPublic:  in.IsPublic,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := item.Validate(); err != nil {
		return nil, apperror.NewInvalidInput("hobby item validation failed", err)
	}
	if err := uc.repo.Save(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

type UpdateHobbyItemInput struct {
	ItemID   uuid.UUID
	OwnerID  uuid.UUID
	Category string
	Title    string
	Status   string
	Rating   int
	Notes    string
	Metadata map[string]any
	IsPublic bool
}

func (uc *HobbyUseCase) UpdateHobbyItem(ctx context.Context, in UpdateHobbyItemInput) (*hobby.HobbyItem, error) {
	item, err := uc.repo.FindByID(ctx, in.ItemID, in.OwnerID)
	if err != nil {
		return nil, err
	}

	item.Category = in.Category
	item.Title = in.Title
	item.Status = in.Status
	item.Rating = in.Rating
	item.Notes = in.Notes
	item.Metadata = in.Metadata
	item.IsPublic = in.IsPublic

	if err := item.Validate(); err != nil {
		return nil, apperror.NewInvalidInput("hobby item validation failed", err)
	}
	if err := uc.repo.Update(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (uc *HobbyUseCase) DeleteHobbyItem(ctx context.Context, id, ownerID uuid.UUID) error {
	return uc.repo.Delete(ctx, id, ownerID)
}

func (uc *HobbyUseCase) GetHobbyItem(ctx context.Context, id, ownerID uuid.UUID) (*hobby.HobbyItem, error) {
	return uc.repo.FindByID(ctx, id, ownerID)
}

func (uc *HobbyUseCase) ListHobbyItems(ctx context.Context, ownerID uuid.UUID, category string, page, limit int) ([]*hobby.HobbyItem, error) {
	if limit <= 0 {
		limit = 20
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit
	return uc.repo.ListByOwnerAndCategory(ctx, ownerID, category, limit, offset)
}

func (uc *HobbyUseCase) ListPublicHobbyItems(ctx context.Context, category string, page, limit int) ([]*hobby.HobbyItem, error) {
	if limit <= 0 {
		limit = 30
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit
	return uc.repo.ListPublicByCategory(ctx, category, limit, offset)
}
