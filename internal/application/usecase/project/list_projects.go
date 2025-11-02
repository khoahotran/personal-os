package project

import (
	"context"

	"github.com/google/uuid"
	"github.com/khoahotran/personal-os/internal/domain/project"
	"github.com/khoahotran/personal-os/pkg/logger"
)

type ListProjectsUseCase struct {
	projectRepo project.Repository
	logger      logger.Logger
}

func NewListProjectsUseCase(pRepo project.Repository, log logger.Logger) *ListProjectsUseCase {
	return &ListProjectsUseCase{projectRepo: pRepo, logger: log}
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
		return nil, err
	}
	return &ListProjectsOutput{Projects: projects}, nil
}

type ListPublicProjectsUseCase struct {
	projectRepo project.Repository
	logger      logger.Logger
}

func NewListPublicProjectsUseCase(pRepo project.Repository, log logger.Logger) *ListPublicProjectsUseCase {
	return &ListPublicProjectsUseCase{projectRepo: pRepo, logger: log}
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
		return nil, err
	}
	return &ListPublicProjectsOutput{Projects: projects}, nil
}
