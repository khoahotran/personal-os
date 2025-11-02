package project

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/khoahotran/personal-os/internal/domain/project"
	"github.com/khoahotran/personal-os/internal/domain/tag"
	"github.com/khoahotran/personal-os/pkg/apperror"
	"github.com/khoahotran/personal-os/pkg/logger"
	"go.uber.org/zap"
)

type UpdateProjectUseCase struct {
	projectRepo project.Repository
	tagRepo     tag.Repository
	logger      logger.Logger
}

func NewUpdateProjectUseCase(pRepo project.Repository, tRepo tag.Repository, log logger.Logger) *UpdateProjectUseCase {
	return &UpdateProjectUseCase{projectRepo: pRepo, tagRepo: tRepo, logger: log}
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
		return nil, apperror.NewInvalidInput("project validation failed", err)
	}
	if err := uc.projectRepo.Update(ctx, p); err != nil {
		return nil, err
	}

	tags, err := uc.tagRepo.FindOrCreateTags(ctx, input.TagNames)
	if err != nil {
		return nil, apperror.NewInternal("failed to process tags", err)
	}
	tagIDs := make([]uuid.UUID, len(tags))
	for i, t := range tags {
		tagIDs[i] = t.ID
	}
	if err = uc.tagRepo.SetTagsForResource(ctx, p.ID, "project", tagIDs); err != nil {
		uc.logger.Warn("Failed to set tags during project update", zap.String("project_id", p.ID.String()), zap.Error(err))
	}

	return &UpdateProjectOutput{Project: p}, nil
}
