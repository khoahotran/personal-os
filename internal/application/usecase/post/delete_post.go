package post

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/khoahotran/personal-os/adapters/event"
	"github.com/khoahotran/personal-os/internal/domain/post"
	"github.com/khoahotran/personal-os/internal/domain/tag"
	"github.com/khoahotran/personal-os/pkg/apperror"
	"github.com/khoahotran/personal-os/pkg/logger"
)

type DeletePostUseCase struct {
	postRepo    post.Repository
	tagRepo     tag.Repository
	kafkaClient *event.KafkaProducerClient
	logger      logger.Logger
}

func NewDeletePostUseCase(pRepo post.Repository, tRepo tag.Repository, kClient *event.KafkaProducerClient, log logger.Logger) *DeletePostUseCase {
	return &DeletePostUseCase{
		postRepo:    pRepo,
		tagRepo:     tRepo,
		kafkaClient: kClient,
		logger:      log,
	}
}

type DeletePostInput struct {
	PostID  uuid.UUID
	OwnerID uuid.UUID
}

func (uc *DeletePostUseCase) Execute(ctx context.Context, input DeletePostInput) error {

	err := uc.tagRepo.SetTagsForResource(ctx, input.PostID, "post", []uuid.UUID{})
	if err != nil {
		return apperror.NewInternal("failed to delete tag relations", err)
	}

	err = uc.postRepo.Delete(ctx, input.PostID, input.OwnerID)
	if err != nil {
		return err
	}

	go func() {
		err := uc.kafkaClient.PublishPostEvent(context.Background(), event.PostEventPayload{
			EventType: event.PostEventTypeDeleted,
			PostID:    input.PostID,
			OwnerID:   input.OwnerID,
		})
		if err != nil {
			uc.logger.Error("Failed to publish Kafka 'deleted' event", err, zap.String("post_id", input.PostID.String()))
		}
	}()

	return nil
}
