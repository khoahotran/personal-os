package media

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/khoahotran/personal-os/adapters/event"
	"github.com/khoahotran/personal-os/internal/application/service"
	"github.com/khoahotran/personal-os/internal/domain/media"
	"github.com/khoahotran/personal-os/pkg/apperror"
	"github.com/khoahotran/personal-os/pkg/logger"
	"go.uber.org/zap"
)

type UploadMediaUseCase struct {
	mediaRepo   media.Repository
	uploader    service.Uploader
	kafkaClient *event.KafkaProducerClient
	logger      logger.Logger
}

func NewUploadMediaUseCase(
	r media.Repository,
	u service.Uploader,
	k *event.KafkaProducerClient,
	log logger.Logger,
) *UploadMediaUseCase {
	return &UploadMediaUseCase{mediaRepo: r, uploader: u, kafkaClient: k, logger: log}
}

type UploadMediaInput struct {
	OwnerID  uuid.UUID
	File     io.Reader
	Metadata map[string]any
	IsPublic bool
	Provider string
}
type UploadMediaOutput struct {
	MediaID uuid.UUID
}

func (uc *UploadMediaUseCase) Execute(ctx context.Context, input UploadMediaInput) (*UploadMediaOutput, error) {
	mediaID := uuid.New()

	originalFolder := fmt.Sprintf("users/%s/media/originals", input.OwnerID.String())
	originalPublicID := mediaID.String()

	originalURL, err := uc.uploader.Upload(ctx, input.File, originalFolder, originalPublicID)
	if err != nil {
		return nil, apperror.NewInternal("failed to upload original media file", err)
	}

	if input.Metadata == nil {
		input.Metadata = make(map[string]any)
	}
	input.Metadata["original_url"] = originalURL
	input.Metadata["original_public_id"] = originalPublicID

	newMedia := &media.Media{
		ID:        mediaID,
		OwnerID:   input.OwnerID,
		Provider:  input.Provider,
		URL:       originalURL,
		Status:    media.StatusPending,
		Metadata:  input.Metadata,
		IsPublic:  input.IsPublic,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if err := uc.mediaRepo.Save(ctx, newMedia); err != nil {
		go uc.uploader.Delete(context.Background(), originalPublicID)
		return nil, err
	}

	go func() {
		payload := event.MediaEventPayload{
			EventType:        event.MediaEventTypeUploaded,
			MediaID:          newMedia.ID,
			OwnerID:          newMedia.OwnerID,
			Provider:         newMedia.Provider,
			OriginalURL:      originalURL,
			OriginalPublicID: originalPublicID,
		}
		if err := uc.kafkaClient.PublishMediaEvent(context.Background(), payload); err != nil {
			uc.logger.Error("Failed to publish Kafka 'media.uploaded' event", err, zap.String("media_id", newMedia.ID.String()))
		}
	}()

	return &UploadMediaOutput{MediaID: mediaID}, nil
}
