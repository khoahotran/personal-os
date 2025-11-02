package project

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/khoahotran/personal-os/internal/domain/project"
	"github.com/khoahotran/personal-os/internal/domain/tag"
)

type GetProjectUseCase struct {
	projectRepo project.Repository
	tagRepo     tag.Repository
}

func NewGetProjectUseCase(pRepo project.Repository, tRepo tag.Repository) *GetProjectUseCase {
	return &GetProjectUseCase{projectRepo: pRepo, tagRepo: tRepo}
}

type GetProjectInput struct {
	ProjectID uuid.UUID
	OwnerID   uuid.UUID
}
type GetProjectOutput struct {
	Project *project.Project
	Tags    []tag.Tag
}

func (uc *GetProjectUseCase) Execute(ctx context.Context, input GetProjectInput) (*GetProjectOutput, error) {
	p, err := uc.projectRepo.FindByID(ctx, input.ProjectID, input.OwnerID)
	if err != nil {
		return nil, err
	}
	fmt.Print("project: ", p)
	tags, err := uc.tagRepo.GetTagsForResource(ctx, p.ID, "project")
	if err != nil {
		fmt.Printf("WARNING: got project %s but failed to get tags: %v\n", p.ID, err)
	}
	return &GetProjectOutput{Project: p, Tags: tags}, nil
}

type GetPublicProjectUseCase struct {
	projectRepo project.Repository
	tagRepo     tag.Repository
}

func NewGetPublicProjectUseCase(pRepo project.Repository, tRepo tag.Repository) *GetPublicProjectUseCase {
	return &GetPublicProjectUseCase{projectRepo: pRepo, tagRepo: tRepo}
}

type GetPublicProjectInput struct {
	Slug string
}
type GetPublicProjectOutput struct {
	Project *project.Project
	Tags    []tag.Tag
}

func (uc *GetPublicProjectUseCase) Execute(ctx context.Context, input GetPublicProjectInput) (*GetPublicProjectOutput, error) {
	p, err := uc.projectRepo.FindPublicBySlug(ctx, input.Slug)
	if err != nil {
		return nil, err
	}
	tags, err := uc.tagRepo.GetTagsForResource(ctx, p.ID, "project")
	if err != nil {
		fmt.Printf("WARNING: got public project %s but failed to get tags: %v\n", p.ID, err)
	}
	return &GetPublicProjectOutput{Project: p, Tags: tags}, nil
}
