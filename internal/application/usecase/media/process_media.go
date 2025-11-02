package media

import (
	"context"
	"errors"

	"github.com/khoahotran/personal-os/adapters/event"
	"github.com/khoahotran/personal-os/internal/application/service"
	"github.com/khoahotran/personal-os/internal/domain/media"
	"github.com/khoahotran/personal-os/pkg/apperror"
	"github.com/khoahotran/personal-os/pkg/logger"
	"go.uber.org/zap"
)

type ProcessMediaUseCase struct {
	mediaRepo media.Repository
	uploader  service.Uploader
	logger    logger.Logger
}

func NewProcessMediaUseCase(r media.Repository, u service.Uploader, log logger.Logger) *ProcessMediaUseCase {
	return &ProcessMediaUseCase{mediaRepo: r, uploader: u, logger: log}
}

func (uc *ProcessMediaUseCase) Execute(ctx context.Context, payload event.MediaEventPayload) error {
	l := uc.logger.With(zap.String("media_id", payload.MediaID.String()), zap.String("event_type", string(payload.EventType)))
	l.Info("Worker UseCase processing media event")

	m, err := uc.mediaRepo.FindByID(ctx, payload.MediaID, payload.OwnerID)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			l.Warn("Media not found, skipping event", zap.String("media_id", payload.MediaID.String()))
			return nil
		}
		return apperror.NewInternal("failed to get media", err)
	}

	if m.Status == media.StatusReady {
		l.Info("Media already in 'ready' state, skipping", zap.String("status", string(m.Status)))
		return nil
	}

	cldClient := uc.uploader.GetClient()
	if cldClient == nil {
		return apperror.NewInternal("could not get cloudinary client from uploader", nil)
	}

	imgAsset, err := cldClient.Image(payload.OriginalPublicID)
	if err != nil {
		return apperror.NewInternal("failed to create cloudinary asset", err)
	}

	imgAsset.Transformation = "c_limit,w_1200"
	mainURLStr, err := imgAsset.String()
	if err != nil {
		return apperror.NewInternal("failed to build main image URL", err)
	}

	imgAsset.Transformation = "c_fill,g_auto,w_400,h_400"
	thumbURLStr, err := imgAsset.String()
	if err != nil {
		return apperror.NewInternal("failed to build thumbnail URL", err)
	}

	l.Info("Generated Cloudinary URLs for media")

	m.URL = mainURLStr
	m.ThumbnailURL = &thumbURLStr
	m.Status = media.StatusReady

	if err := uc.mediaRepo.Update(ctx, m); err != nil {
		return apperror.NewInternal("failed to update media to 'ready'", err)
	}

	l.Info("Successfully processed media", zap.String("status", string(m.Status)))
	return nil
}
