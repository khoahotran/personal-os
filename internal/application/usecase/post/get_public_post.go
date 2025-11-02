package post

import (
	"context"

	"github.com/khoahotran/personal-os/internal/domain/post"
	"github.com/khoahotran/personal-os/internal/domain/tag"
	"github.com/khoahotran/personal-os/pkg/logger"
	"go.uber.org/zap"
)

type GetPublicPostUseCase struct {
	postRepo post.Repository
	tagRepo  tag.Repository
	logger   logger.Logger
}

func NewGetPublicPostUseCase(pRepo post.Repository, tRepo tag.Repository, log logger.Logger) *GetPublicPostUseCase {
	return &GetPublicPostUseCase{
		postRepo: pRepo,
		tagRepo:  tRepo,
		logger:   log,
	}
}

type GetPublicPostInput struct {
	Slug string
}

type GetPublicPostOutput struct {
	Post *post.Post
	Tags []tag.Tag
}

func (uc *GetPublicPostUseCase) Execute(ctx context.Context, input GetPublicPostInput) (*GetPublicPostOutput, error) {
	p, err := uc.postRepo.FindPublicBySlug(ctx, input.Slug)
	if err != nil {
		return nil, err
	}

	tags, err := uc.tagRepo.GetTagsForResource(ctx, p.ID, "post")
	if err != nil {
		uc.logger.Warn("Failed to get tags for public post", zap.String("post_id", p.ID.String()), zap.Error(err))
	}

	return &GetPublicPostOutput{
		Post: p,
		Tags: tags,
	}, nil
}
