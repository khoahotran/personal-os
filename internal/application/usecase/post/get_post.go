package post

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/khoahotran/personal-os/internal/domain/post"
	"github.com/khoahotran/personal-os/internal/domain/tag"
)

type GetPostUseCase struct {
	postRepo post.Repository
	tagRepo  tag.Repository
}

func NewGetPostUseCase(pRepo post.Repository, tRepo tag.Repository) *GetPostUseCase {
	return &GetPostUseCase{
		postRepo: pRepo,
		tagRepo:  tRepo,
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
		return nil, fmt.Errorf("get post failed: %w", err)
	}

	tags, err := uc.tagRepo.GetTagsForResource(ctx, p.ID, "post")
	if err != nil {
		fmt.Printf("WARNING: got post %s but get tags failed: %v\n", p.ID, err)
	}

	return &GetPostOutput{
		Post: p,
		Tags: tags,
	}, nil
}
