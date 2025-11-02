package post

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/khoahotran/personal-os/adapters/event"
	"github.com/khoahotran/personal-os/internal/domain/post"
	"github.com/khoahotran/personal-os/internal/domain/tag"
	"github.com/khoahotran/personal-os/pkg/apperror"
	"github.com/khoahotran/personal-os/pkg/logger"
)

type UpdatePostUseCase struct {
	postRepo    post.Repository
	tagRepo     tag.Repository
	kafkaClient *event.KafkaProducerClient
	logger      logger.Logger
}

func NewUpdatePostUseCase(pRepo post.Repository, tRepo tag.Repository, kClient *event.KafkaProducerClient, log logger.Logger) *UpdatePostUseCase {
	return &UpdatePostUseCase{
		postRepo:    pRepo,
		tagRepo:     tRepo,
		kafkaClient: kClient,
		logger:      log,
	}
}

type UpdatePostInput struct {
	PostID  uuid.UUID
	OwnerID uuid.UUID

	Title   string
	Content string
	Slug    string
	Status  post.PostStatus
	Tags    []string
}

type UpdatePostOutput struct {
	Post *post.Post
}

func (uc *UpdatePostUseCase) Execute(ctx context.Context, input UpdatePostInput) (*UpdatePostOutput, error) {
	existingPost, err := uc.postRepo.FindByID(ctx, input.PostID, input.OwnerID)
	if err != nil {
		return nil, err
	}

	if existingPost.ContentMarkdown != input.Content {
		existingPost.AddVersion(existingPost.UpdatedAt, existingPost.ContentMarkdown)
	}

	existingPost.Title = input.Title
	existingPost.ContentMarkdown = input.Content
	existingPost.Slug = input.Slug
	existingPost.Status = input.Status
	existingPost.UpdatedAt = time.Now().UTC()

	if err := existingPost.Validate(); err != nil {
		return nil, apperror.NewInvalidInput("validation failed", err)
	}

	if err := uc.postRepo.Update(ctx, existingPost); err != nil {
		return nil, err
	}

	tags, err := uc.tagRepo.FindOrCreateTags(ctx, input.Tags)
	if err != nil {
		return nil, apperror.NewInternal("failed to process tags", err)
	}

	tagIDs := make([]uuid.UUID, len(tags))
	for i, t := range tags {
		tagIDs[i] = t.ID
	}

	err = uc.tagRepo.SetTagsForResource(ctx, existingPost.ID, "post", tagIDs)
	if err != nil {
		uc.logger.Warn("Failed to set tags during post update", zap.String("post_id", existingPost.ID.String()), zap.Error(err))
	}

	go func() {
		eventType := event.PostEventTypeUpdated
		if existingPost.Status == post.StatusPublic {
			eventType = event.PostEventTypePublished
		}

		err := uc.kafkaClient.PublishPostEvent(context.Background(), event.PostEventPayload{
			EventType: eventType,
			PostID:    existingPost.ID,
			OwnerID:   existingPost.OwnerID,
		})
		if err != nil {
			uc.logger.Error("Failed to publish Kafka 'updated' event", err, zap.String("post_id", existingPost.ID.String()))
		}
	}()

	return &UpdatePostOutput{Post: existingPost}, nil
}
