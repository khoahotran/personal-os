package post

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/khoahotran/personal-os/adapters/event"
	"github.com/khoahotran/personal-os/internal/domain/post"
	"github.com/khoahotran/personal-os/internal/domain/tag"
)

type UpdatePostUseCase struct {
	postRepo    post.Repository
	tagRepo     tag.Repository
	kafkaClient *event.KafkaProducerClient
}

func NewUpdatePostUseCase(pRepo post.Repository, tRepo tag.Repository, kClient *event.KafkaProducerClient) *UpdatePostUseCase {
	return &UpdatePostUseCase{
		postRepo:    pRepo,
		tagRepo:     tRepo,
		kafkaClient: kClient,
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
		return nil, err
	}

	if err := uc.postRepo.Update(ctx, existingPost); err != nil {
		return nil, fmt.Errorf("update post failed: %w", err)
	}

	tags, err := uc.tagRepo.FindOrCreateTags(ctx, input.Tags)
	if err != nil {
		return nil, fmt.Errorf("process tags failed: %w", err)
	}

	tagIDs := make([]uuid.UUID, len(tags))
	for i, t := range tags {
		tagIDs[i] = t.ID
	}

	err = uc.tagRepo.SetTagsForResource(ctx, existingPost.ID, "post", tagIDs)
	if err != nil {

		fmt.Printf("WARNING: update post %s but assign tag failed: %v\n", existingPost.ID, err)
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
			log.Printf("ERROR (background): Send event to Kafka failed for post %s: %v", existingPost.ID, err)
		}
	}()

	return &UpdatePostOutput{Post: existingPost}, nil
}
