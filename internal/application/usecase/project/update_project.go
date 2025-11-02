package project

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/khoahotran/personal-os/internal/domain/project"
	"github.com/khoahotran/personal-os/internal/domain/tag"
)

type UpdateProjectUseCase struct {
	projectRepo project.Repository
	tagRepo     tag.Repository
}

func NewUpdateProjectUseCase(pRepo project.Repository, tRepo tag.Repository) *UpdateProjectUseCase {
	return &UpdateProjectUseCase{projectRepo: pRepo, tagRepo: tRepo}
}

type UpdateProjectInput struct {
	ProjectID     uuid.UUID
	OwnerID       uuid.UUID
	Title         string
	Slug          string
	Description   string
	Stack         []string
	RepositoryURL *string
	LiveURL       *string
	Media         []project.ProjectMedia
	IsPublic      bool
	TagNames      []string
}
type UpdateProjectOutput struct {
	Project *project.Project
}

func (uc *UpdateProjectUseCase) Execute(ctx context.Context, input UpdateProjectInput) (*UpdateProjectOutput, error) {

	p, err := uc.projectRepo.FindByID(ctx, input.ProjectID, input.OwnerID)
	if err != nil {
		return nil, err
	}

	p.Title = input.Title
	p.Slug = input.Slug
	p.Description = input.Description
	p.Stack = input.Stack
	p.RepositoryURL = input.RepositoryURL
	p.LiveURL = input.LiveURL
	p.Media = input.Media
	p.IsPublic = input.IsPublic
	p.UpdatedAt = time.Now().UTC()

	if err := p.Validate(); err != nil {
		return nil, err
	}
	if err := uc.projectRepo.Update(ctx, p); err != nil {
		return nil, fmt.Errorf("update project failed: %w", err)
	}

	tags, err := uc.tagRepo.FindOrCreateTags(ctx, input.TagNames)
	if err != nil {

		return nil, fmt.Errorf("process tags failed: %w", err)
	}
	tagIDs := make([]uuid.UUID, len(tags))
	for i, t := range tags {
		tagIDs[i] = t.ID
	}
	if err = uc.tagRepo.SetTagsForResource(ctx, p.ID, "project", tagIDs); err != nil {
		fmt.Printf("WARNING: updated project %s but set tags failed: %v\n", p.ID, err)
	}

	return &UpdateProjectOutput{Project: p}, nil
}
