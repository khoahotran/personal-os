package media

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/khoahotran/personal-os/adapters/event"
	"github.com/khoahotran/personal-os/internal/application/service"
	"github.com/khoahotran/personal-os/internal/domain/media"
)

type UploadMediaUseCase struct {
	mediaRepo   media.Repository
	uploader    service.Uploader
	kafkaClient *event.KafkaProducerClient
}

func NewUploadMediaUseCase(r media.Repository, u service.Uploader, k *event.KafkaProducerClient) *UploadMediaUseCase {
	return &UploadMediaUseCase{mediaRepo: r, uploader: u, kafkaClient: k}
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
	now := time.Now().UTC()
	mediaID := uuid.New()

	originalFolder := fmt.Sprintf("users/%s/media/originals", input.OwnerID.String())
	originalPublicID := mediaID.String()

	originalURL, err := uc.uploader.Upload(ctx, input.File, originalFolder, originalPublicID)
	if err != nil {
		return nil, fmt.Errorf("upload original file media failed: %w", err)
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
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := uc.mediaRepo.Save(ctx, newMedia); err != nil {
		go uc.uploader.Delete(context.Background(), originalPublicID)
		return nil, fmt.Errorf("save media metadata failed: %w", err)
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
			log.Printf("ERROR (background): Send 'media.uploaded' event to Kafka failed: %v", err)
		}
	}()

	return &UploadMediaOutput{MediaID: mediaID}, nil
}
