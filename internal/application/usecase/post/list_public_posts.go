package post

import (
	"context"

	"github.com/khoahotran/personal-os/internal/domain/post"
	"github.com/khoahotran/personal-os/internal/domain/tag"
	"github.com/khoahotran/personal-os/pkg/logger"
)

type ListPublicPostsUseCase struct {
	postRepo post.Repository
	tagRepo  tag.Repository
	logger   logger.Logger
}

func NewListPublicPostsUseCase(pRepo post.Repository, tRepo tag.Repository, log logger.Logger) *ListPublicPostsUseCase {
	return &ListPublicPostsUseCase{
		postRepo: pRepo,
		tagRepo:  tRepo,
		logger:   log,
	}
}

type ListPublicPostsInput struct {
	Page  int
	Limit int
}

type ListPublicPostsOutput struct {
	Posts []*post.Post
}

func (uc *ListPublicPostsUseCase) Execute(ctx context.Context, input ListPublicPostsInput) (*ListPublicPostsOutput, error) {

	if input.Limit <= 0 || input.Limit > 100 {
		input.Limit = 10
	}
	if input.Page <= 0 {
		input.Page = 1
	}
	offset := (input.Page - 1) * input.Limit

	posts, err := uc.postRepo.ListPublic(ctx, input.Limit, offset)
	if err != nil {
		return nil, err
	}

	return &ListPublicPostsOutput{Posts: posts}, nil
}
