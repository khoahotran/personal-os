package post

import (
	"context"
	"errors"

	"github.com/khoahotran/personal-os/adapters/event"
	"github.com/khoahotran/personal-os/internal/application/service"
	"github.com/khoahotran/personal-os/internal/domain/post"
	"github.com/khoahotran/personal-os/pkg/apperror"
	"github.com/khoahotran/personal-os/pkg/logger"
	"go.uber.org/zap"
)

type ProcessPostEventUseCase struct {
	postRepo post.Repository
	uploader service.Uploader
	embedder service.EmbeddingService
	logger   logger.Logger
}

func NewProcessPostEventUseCase(pr post.Repository, up service.Uploader, em service.EmbeddingService, log logger.Logger) *ProcessPostEventUseCase {
	return &ProcessPostEventUseCase{postRepo: pr, uploader: up, embedder: em, logger: log}
}

func (uc *ProcessPostEventUseCase) Execute(ctx context.Context, payload event.PostEventPayload) error {
	l := uc.logger.With(zap.String("post_id", payload.PostID.String()), zap.String("event_type", string(payload.EventType)))
	l.Info("Worker UseCase processing event")
	p, err := uc.postRepo.FindByID(ctx, payload.PostID, payload.OwnerID)
	if err != nil {
		if errors.Is(err, post.ErrPostNotFound) {
			l.Warn("Post not found, skipping event", zap.String("post_id", payload.PostID.String()))
			return nil
		}
		return apperror.NewInternal("failed to get post", err)
	}

	if p.Status != post.StatusPending {
		l.Info("Post not in 'pending' state, skipping", zap.String("status", string(p.Status)))
		return nil
	}

	cldClient := uc.uploader.GetClient()
	if cldClient == nil {
		return apperror.NewInternal("could not get cloudinary client from uploader", nil)
	}

	originalPublicID, ok := p.Metadata["original_public_id"].(string)
	l.Info("[DEBUG]: ", zap.String("original_public_id", originalPublicID))
	if !ok || originalPublicID == "" {
		return apperror.NewInvalidInput("original_public_id not found in metadata", nil)
	}

	// Get original image asset
	imgAsset, err := cldClient.Image(originalPublicID)
	if err != nil {
		return apperror.NewInternal("failed to create cloudinary asset", err)
	}

	// OG image transform
	imgAsset.Transformation = "c_fill,g_auto,w_1200,h_630"
	ogImageURL, err := imgAsset.String()
	if err != nil {
		return apperror.NewInternal("failed to build OG image URL", err)
	}

	// Thumbnail transform
	imgAsset.Transformation = "c_limit,w_400"
	thumbURL, err := imgAsset.String()
	if err != nil {
		return apperror.NewInternal("failed to build Thumbnail URL", err)
	}

	if payload.EventType == event.PostEventTypeCreated || payload.EventType == event.PostEventTypeUpdated {
		l.Info("Generating embeddings for post content...")
		embedding, err := uc.embedder.GenerateEmbeddings(ctx, p.ContentMarkdown)
		if err != nil {
			return apperror.NewInternal("failed to generate embeddings", err)
		}
		p.Embedding = embedding
		l.Info("Embeddings generated successfully")
	}

	requestedStatusStr, _ := p.Metadata["requested_status"].(string)
	if requestedStatusStr == "" {
		l.Warn("request status violate: %s, fallback to draft\n", zap.String("request_status", requestedStatusStr))
		requestedStatusStr = string(post.StatusDraft)
	}
	p.Status = post.PostStatus(requestedStatusStr)
	l.Info("[DEBUG]: ", zap.String("ogImageURL", ogImageURL))
	l.Info("[DEBUG]: ", zap.String("thumbURL", thumbURL))

	p.MarkAsReady(ogImageURL, thumbURL)

	if err := uc.postRepo.Update(ctx, p); err != nil {
		return apperror.NewInternal("failed to update post with OG image", err)
	}

	l.Info("Successfully updated Post with OG Image", zap.String("status", string(p.Status)))
	return nil
}
