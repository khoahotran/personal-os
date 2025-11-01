package post

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"

	"github.com/khoahotran/personal-os/adapters/event"
	"github.com/khoahotran/personal-os/internal/domain/post"
	"github.com/khoahotran/personal-os/internal/domain/tag"
)

type DeletePostUseCase struct {
	postRepo    post.Repository
	tagRepo     tag.Repository
	kafkaClient *event.KafkaProducerClient
}

func NewDeletePostUseCase(pRepo post.Repository, tRepo tag.Repository, kClient *event.KafkaProducerClient) *DeletePostUseCase {
	return &DeletePostUseCase{
		postRepo:    pRepo,
		tagRepo:     tRepo,
		kafkaClient: kClient,
	}
}

type DeletePostInput struct {
	PostID  uuid.UUID
	OwnerID uuid.UUID
}

func (uc *DeletePostUseCase) Execute(ctx context.Context, input DeletePostInput) error {

	err := uc.tagRepo.SetTagsForResource(ctx, input.PostID, "post", []uuid.UUID{})
	if err != nil {
		return fmt.Errorf("delete tag links failed: %w", err)
	}

	err = uc.postRepo.Delete(ctx, input.PostID, input.OwnerID)
	if err != nil {
		return fmt.Errorf("delete post failed: %w", err)
	}

	go func() {
		err := uc.kafkaClient.PublishPostEvent(context.Background(), event.PostEventPayload{
			EventType: event.PostEventTypeDeleted,
			PostID:    input.PostID,
			OwnerID:   input.OwnerID,
		})
		if err != nil {
			log.Printf("ERROR (background): Sent event to Kafka failed for post %s: %v", input.PostID, err)
		}
	}()

	return nil
}
