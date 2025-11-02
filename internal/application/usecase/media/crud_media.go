package media

import (
	"context"

	"github.com/google/uuid"
	"github.com/khoahotran/personal-os/internal/application/service"
	"github.com/khoahotran/personal-os/internal/domain/media"
	"github.com/khoahotran/personal-os/pkg/logger"
	"go.uber.org/zap"
)

type ListPublicMediaUseCase struct {
	mediaRepo media.Repository
	logger    logger.Logger
}

func NewListPublicMediaUseCase(r media.Repository, log logger.Logger) *ListPublicMediaUseCase {
	return &ListPublicMediaUseCase{mediaRepo: r, logger: log}
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
		return nil, err
	}
	return &ListPublicMediaOutput{Medias: medias}, nil
}

type UpdateMediaUseCase struct {
	mediaRepo media.Repository
	logger    logger.Logger
}

func NewUpdateMediaUseCase(r media.Repository, log logger.Logger) *UpdateMediaUseCase {
	return &UpdateMediaUseCase{mediaRepo: r, logger: log}
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

	if err := uc.mediaRepo.Update(ctx, m); err != nil {
		return err
	}
	return nil
}

type DeleteMediaUseCase struct {
	mediaRepo media.Repository
	uploader  service.Uploader
	logger    logger.Logger
}

func NewDeleteMediaUseCase(r media.Repository, u service.Uploader, log logger.Logger) *DeleteMediaUseCase {
	return &DeleteMediaUseCase{mediaRepo: r, uploader: u, logger: log}
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
			uc.logger.Warn("Failed to delete media from Cloudinary, proceeding with DB delete", zap.String("public_id", publicID), zap.Error(err))
		}
	} else {
		uc.logger.Warn("No 'original_public_id' found in metadata, cannot delete from Cloudinary", zap.String("media_id", m.ID.String()))
	}

	return uc.mediaRepo.Delete(ctx, in.MediaID, in.OwnerID)
}
