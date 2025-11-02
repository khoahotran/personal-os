package post

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/khoahotran/personal-os/adapters/event"
	"github.com/khoahotran/personal-os/internal/application/service"
	"github.com/khoahotran/personal-os/internal/domain/post"
	"github.com/khoahotran/personal-os/internal/domain/tag"
	"github.com/khoahotran/personal-os/pkg/apperror"
	"github.com/khoahotran/personal-os/pkg/logger"
)

type CreatePostUseCase struct {
	postRepo    post.Repository
	tagRepo     tag.Repository
	kafkaClient *event.KafkaProducerClient
	uploader    service.Uploader
	logger      logger.Logger
}

func NewCreatePostUseCase(pRepo post.Repository, tRepo tag.Repository, kClient *event.KafkaProducerClient, uploader service.Uploader, log logger.Logger) *CreatePostUseCase {
	return &CreatePostUseCase{
		postRepo:    pRepo,
		tagRepo:     tRepo,
		kafkaClient: kClient,
		uploader:    uploader,
		logger:      log,
	}
}

type CreatePostInput struct {
	OwnerID         uuid.UUID
	Title           string
	Content         string
	Slug            string
	RequestedStatus post.PostStatus
	TagNames        []string
	File            io.Reader
	Metadata        map[string]any
}

type CreatePostOutput struct {
	PostID uuid.UUID
	Slug   string
}

func (uc *CreatePostUseCase) Execute(ctx context.Context, input CreatePostInput) (*CreatePostOutput, error) {
	if input.Slug == "" {
		input.Slug = strings.ToLower(strings.ReplaceAll(input.Title, " ", "-"))
	}

	now := time.Now().UTC()

	if input.Metadata == nil {
		input.Metadata = make(map[string]any)
	}
	input.Metadata["requested_status"] = input.RequestedStatus

	newPost := &post.Post{
		ID:              uuid.New(),
		OwnerID:         input.OwnerID,
		Slug:            input.Slug,
		Title:           input.Title,
		ContentMarkdown: input.Content,
		Status:          post.StatusPending,
		Metadata:        input.Metadata,
		VersionHistory:  []post.PostVersion{},
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := newPost.Validate(); err != nil {
		return nil, apperror.NewInvalidInput("validation failed", err)
	}

	originalFolder := fmt.Sprintf("users/%s/originals", input.OwnerID.String())
	originalPublicID := newPost.ID.String()

	originalURL, err := uc.uploader.Upload(ctx, input.File, originalFolder, originalPublicID)
	if err != nil {
		return nil, apperror.NewInternal("failed to upload original file", err)
	}
	originalPublicID =  originalFolder + newPost.ID.String()

	if newPost.Metadata == nil {
		newPost.Metadata = make(map[string]any)
	}
	newPost.Metadata["original_url"] = originalURL
	newPost.Metadata["original_public_id"] = originalPublicID

	tags, err := uc.tagRepo.FindOrCreateTags(ctx, input.TagNames)
	if err != nil {
		return nil, fmt.Errorf("process tags failed: %w", err)
	}

	if err := uc.postRepo.Save(ctx, newPost); err != nil {
		go uc.uploader.Delete(context.Background(), originalPublicID)
		return nil, err
	}

	tagIDs := make([]uuid.UUID, len(tags))
	for i, t := range tags {
		tagIDs[i] = t.ID
	}

	if err = uc.tagRepo.SetTagsForResource(ctx, newPost.ID, "post", tagIDs); err != nil {
		uc.logger.Warn("Failed to set tags for new post", zap.String("post_id", newPost.ID.String()), zap.Error(err))
	}

	go func() {
		err := uc.kafkaClient.PublishPostEvent(context.Background(), event.PostEventPayload{
			EventType: event.PostEventTypeCreated,
			PostID:    newPost.ID,
			OwnerID:   newPost.OwnerID,
		})
		if err != nil {
			uc.logger.Error("Failed to publish Kafka 'created' event", err, zap.String("post_id", newPost.ID.String()))
		}
	}()

	return &CreatePostOutput{
		PostID: newPost.ID,
		Slug:   newPost.Slug,
	}, nil
}
