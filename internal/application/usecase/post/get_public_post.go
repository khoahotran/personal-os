package post

import (
	"context"
	"fmt"

	"github.com/khoahotran/personal-os/internal/domain/post"
	"github.com/khoahotran/personal-os/internal/domain/tag"
)

type GetPublicPostUseCase struct {
	postRepo post.Repository
	tagRepo  tag.Repository
}

func NewGetPublicPostUseCase(pRepo post.Repository, tRepo tag.Repository) *GetPublicPostUseCase {
	return &GetPublicPostUseCase{
		postRepo: pRepo,
		tagRepo:  tRepo,
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

		return nil, fmt.Errorf("get public post failed: %w", err)
	}

	tags, err := uc.tagRepo.GetTagsForResource(ctx, p.ID, "post")
	if err != nil {

		fmt.Printf("WARNING: got public post %s but get tags failed: %v\n", p.ID, err)
	}

	return &GetPublicPostOutput{
		Post: p,
		Tags: tags,
	}, nil
}
