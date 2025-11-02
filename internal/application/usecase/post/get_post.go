package post

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/khoahotran/personal-os/internal/domain/post"
	"github.com/khoahotran/personal-os/internal/domain/tag"
	"github.com/khoahotran/personal-os/pkg/logger"
)

type GetPostUseCase struct {
	postRepo post.Repository
	tagRepo  tag.Repository
	logger   logger.Logger
}

func NewGetPostUseCase(pRepo post.Repository, tRepo tag.Repository, log logger.Logger) *GetPostUseCase {
	return &GetPostUseCase{
		postRepo: pRepo,
		tagRepo:  tRepo,
		logger:   log,
	}
}

type GetPostInput struct {
	PostID  uuid.UUID
	OwnerID uuid.UUID
}

type GetPostOutput struct {
	Post *post.Post
	Tags []tag.Tag
}

func (uc *GetPostUseCase) Execute(ctx context.Context, input GetPostInput) (*GetPostOutput, error) {
	p, err := uc.postRepo.FindByID(ctx, input.PostID, input.OwnerID)
	if err != nil {
		return nil, err
	}

	tags, err := uc.tagRepo.GetTagsForResource(ctx, p.ID, "post")
	if err != nil {
		uc.logger.Warn("Failed to get tags for post", zap.String("post_id", p.ID.String()), zap.Error(err))
	}

	return &GetPostOutput{
		Post: p,
		Tags: tags,
	}, nil
}
