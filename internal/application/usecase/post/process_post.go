package post

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/khoahotran/personal-os/adapters/event"
	"github.com/khoahotran/personal-os/internal/application/service"
	"github.com/khoahotran/personal-os/internal/domain/post"
)

type ProcessPostEventUseCase struct {
	postRepo post.Repository
	uploader service.Uploader
}

func NewProcessPostEventUseCase(pr post.Repository, up service.Uploader) *ProcessPostEventUseCase {
	return &ProcessPostEventUseCase{postRepo: pr, uploader: up}
}

func (uc *ProcessPostEventUseCase) Execute(ctx context.Context, payload event.PostEventPayload) error {
	log.Printf("Worker UseCase processing event: %s for PostID: %s", payload.EventType, payload.PostID)

	p, err := uc.postRepo.FindByID(ctx, payload.PostID, payload.OwnerID)
	if err != nil {
		if errors.Is(err, post.ErrPostNotFound) {
			log.Printf("WARN: Post %s not found, skip.", payload.PostID)
			return nil
		}
		return fmt.Errorf("get post failed: %w", err)
	}

	if p.Status != post.StatusPending {
		log.Printf("INFO: Post %s status != 'pending' (processed?), skip.", p.ID)
		return nil
	}

	cldClient := uc.uploader.GetClient()
	if cldClient == nil {
		return fmt.Errorf("get cloudinary client from uploader failed")
	}

	originalPublicID := fmt.Sprintf("users/%s/originals/%s", payload.OwnerID.String(), payload.PostID.String())

	// Get original image asset
	imgAsset, err := cldClient.Image(originalPublicID)
	if err != nil {
		return fmt.Errorf("init cloudinary asset failed: %w", err)
	}

	// OG image transform
	imgAsset.Transformation = "c_fill,g_auto,w_1200,h_630"
	ogImageURL, err := imgAsset.String()
	if err != nil {
		return fmt.Errorf("build OG image URL failed: %w", err)
	}

	// Thumbnail transform
	imgAsset.Transformation = "c_limit,w_400"
	thumbURL, err := imgAsset.String()
	if err != nil {
		return fmt.Errorf("build Thumbnail URL failed: %w", err)
	}

	requestedStatusStr, _ := p.Metadata["requested_status"].(string)
	if requestedStatusStr == "" {
		fmt.Printf("request status violate: %s, fallback to draft\n", requestedStatusStr)
		requestedStatusStr = string(post.StatusDraft)
	}
	p.Status = post.PostStatus(requestedStatusStr)

	p.MarkAsReady(ogImageURL, thumbURL)

	if err := uc.postRepo.Update(ctx, p); err != nil {
		return fmt.Errorf("update post failed %s with OG image and thumbnail: %w", p.ID, err)
	}

	log.Printf("Updated post %s with OG Image, thumbnail and status '%s'.", p.ID, p.Status)
	return nil
}
