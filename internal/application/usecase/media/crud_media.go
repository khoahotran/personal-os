package media

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/khoahotran/personal-os/internal/application/service"
	"github.com/khoahotran/personal-os/internal/domain/media"
)

// List public

type ListPublicMediaUseCase struct {
	mediaRepo media.Repository
}

func NewListPublicMediaUseCase(r media.Repository) *ListPublicMediaUseCase {
	return &ListPublicMediaUseCase{mediaRepo: r}
}

type ListPublicMediaInput struct{ Limit, Offset int }
type ListPublicMediaOutput struct{ Medias []*media.Media }

func (uc *ListPublicMediaUseCase) Execute(ctx context.Context, in ListPublicMediaInput) (*ListPublicMediaOutput, error) {
	if in.Limit <= 0 {
		in.Limit = 30
	}
	if in.Offset < 0 {
		in.Offset = 0
	}
	medias, err := uc.mediaRepo.ListPublic(ctx, in.Limit, in.Offset)
	if err != nil {
		return nil, fmt.Errorf("get public media failed: %w", err)
	}
	return &ListPublicMediaOutput{Medias: medias}, nil
}

// Update

type UpdateMediaUseCase struct {
	mediaRepo media.Repository
}

func NewUpdateMediaUseCase(r media.Repository) *UpdateMediaUseCase {
	return &UpdateMediaUseCase{mediaRepo: r}
}

type UpdateMediaInput struct {
	OwnerID  uuid.UUID
	MediaID  uuid.UUID
	Metadata map[string]any
	IsPublic bool
}

func (uc *UpdateMediaUseCase) Execute(ctx context.Context, in UpdateMediaInput) error {
	m, err := uc.mediaRepo.FindByID(ctx, in.MediaID, in.OwnerID)
	if err != nil {
		return err
	}
	m.Metadata = in.Metadata
	m.IsPublic = in.IsPublic
	return uc.mediaRepo.Update(ctx, m)
}

// Delete

type DeleteMediaUseCase struct {
	mediaRepo media.Repository
	uploader  service.Uploader
}

func NewDeleteMediaUseCase(r media.Repository, u service.Uploader) *DeleteMediaUseCase {
	return &DeleteMediaUseCase{mediaRepo: r, uploader: u}
}

type DeleteMediaInput struct {
	OwnerID uuid.UUID
	MediaID uuid.UUID
}

func (uc *DeleteMediaUseCase) Execute(ctx context.Context, in DeleteMediaInput) error {
	m, err := uc.mediaRepo.FindByID(ctx, in.MediaID, in.OwnerID)
	if err != nil {
		return err
	}

	if publicID, ok := m.Metadata["original_public_id"].(string); ok {
		if err := uc.uploader.Delete(ctx, publicID); err != nil {
			log.Printf("WARNING: delete media %s failed (Cloudinary): %v", m.ID, err)

		}
	} else {
		log.Printf("WARNING: 'original_public_id' not found for media %s", m.ID)
	}

	return uc.mediaRepo.Delete(ctx, in.MediaID, in.OwnerID)
}
