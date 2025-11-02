package post

import (
	"context"

	"github.com/google/uuid"

	"github.com/khoahotran/personal-os/internal/domain/post"
	"github.com/khoahotran/personal-os/internal/domain/tag"
	"github.com/khoahotran/personal-os/pkg/logger"
)

type ListPostsUseCase struct {
	postRepo post.Repository
	tagRepo  tag.Repository
	logger   logger.Logger
}

func NewListPostsUseCase(pRepo post.Repository, tRepo tag.Repository, log logger.Logger) *ListPostsUseCase {
	return &ListPostsUseCase{
		postRepo: pRepo,
		tagRepo:  tRepo,
		logger:   log,
	}
}

type ListPostsInput struct {
	OwnerID uuid.UUID
	Page    int
	Limit   int
}

type ListPostsOutput struct {
	Posts []*post.Post
}

func (uc *ListPostsUseCase) Execute(ctx context.Context, input ListPostsInput) (*ListPostsOutput, error) {

	if input.Limit <= 0 || input.Limit > 100 {
		input.Limit = 20
	}
	if input.Page <= 0 {
		input.Page = 1
	}
	offset := (input.Page - 1) * input.Limit

	posts, err := uc.postRepo.ListByOwner(ctx, input.OwnerID, input.Limit, offset)
	if err != nil {
		return nil, err
	}

	return &ListPostsOutput{Posts: posts}, nil
}
