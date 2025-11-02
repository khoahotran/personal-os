package media

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/khoahotran/personal-os/adapters/event"
	"github.com/khoahotran/personal-os/internal/application/service"
	"github.com/khoahotran/personal-os/internal/domain/media"
)

type ProcessMediaUseCase struct {
	mediaRepo media.Repository
	uploader  service.Uploader
}

func NewProcessMediaUseCase(r media.Repository, u service.Uploader) *ProcessMediaUseCase {
	return &ProcessMediaUseCase{mediaRepo: r, uploader: u}
}

func (uc *ProcessMediaUseCase) Execute(ctx context.Context, payload event.MediaEventPayload) error {
	log.Printf("Worker UseCase processing event: %s cho MediaID: %s", payload.EventType, payload.MediaID)

	m, err := uc.mediaRepo.FindByID(ctx, payload.MediaID, payload.OwnerID)
	if err != nil {
		if errors.Is(err, errors.New("media not found")) {
			log.Printf("WARN: Media %s not found, skip.", payload.MediaID)
			return nil
		}
		return fmt.Errorf("get media failed: %w", err)
	}

	if m.Status == media.StatusReady {
		log.Printf("INFO: Media %s has status is 'ready', skip.", m.ID)
		return nil
	}

	cldClient := uc.uploader.GetClient()
	if cldClient == nil {
		return fmt.Errorf("get cloudinary client from uploader failed")
	}

	// Get original image asset
	imgAsset, err := cldClient.Image(payload.OriginalPublicID)
	if err != nil {
		return fmt.Errorf("init cloudinary asset failed: %w", err)
	}

	// Thumbnail transform
	imgAsset.Transformation = "c_limit,w_400"
	thumbURL, err := imgAsset.String()
	if err != nil {
		return fmt.Errorf("build Thumbnail URL failed: %w", err)
	}

	m.URL = payload.OriginalURL
	m.ThumbnailURL = &thumbURL
	m.Status = media.StatusReady

	if err := uc.mediaRepo.Update(ctx, m); err != nil {
		return fmt.Errorf("update media %s to 'ready' failed: %w", m.ID, err)
	}

	log.Printf("Processed Media %s (Thumbnail: %s)", m.ID, thumbURL)
	return nil
}
