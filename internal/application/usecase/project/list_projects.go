package project

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/khoahotran/personal-os/internal/domain/project"
)

type ListProjectsUseCase struct {
	projectRepo project.Repository
}

func NewListProjectsUseCase(pRepo project.Repository) *ListProjectsUseCase {
	return &ListProjectsUseCase{projectRepo: pRepo}
}

type ListProjectsInput struct {
	OwnerID uuid.UUID
	Page    int
	Limit   int
}
type ListProjectsOutput struct {
	Projects []*project.Project
}

func (uc *ListProjectsUseCase) Execute(ctx context.Context, input ListProjectsInput) (*ListProjectsOutput, error) {
	if input.Limit <= 0 {
		input.Limit = 20
	}
	if input.Page <= 0 {
		input.Page = 1
	}
	offset := (input.Page - 1) * input.Limit

	projects, err := uc.projectRepo.ListByOwner(ctx, input.OwnerID, input.Limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get list projects failed: %w", err)
	}
	return &ListProjectsOutput{Projects: projects}, nil
}

type ListPublicProjectsUseCase struct {
	projectRepo project.Repository
}

func NewListPublicProjectsUseCase(pRepo project.Repository) *ListPublicProjectsUseCase {
	return &ListPublicProjectsUseCase{projectRepo: pRepo}
}

type ListPublicProjectsInput struct {
	Page  int
	Limit int
}
type ListPublicProjectsOutput struct {
	Projects []*project.Project
}

func (uc *ListPublicProjectsUseCase) Execute(ctx context.Context, input ListPublicProjectsInput) (*ListPublicProjectsOutput, error) {
	if input.Limit <= 0 {
		input.Limit = 10
	}
	if input.Page <= 0 {
		input.Page = 1
	}
	offset := (input.Page - 1) * input.Limit

	projects, err := uc.projectRepo.ListPublic(ctx, input.Limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get list public projects failed: %w", err)
	}
	return &ListPublicProjectsOutput{Projects: projects}, nil
}
