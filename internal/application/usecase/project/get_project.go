package project

import (
	"context"

	"github.com/google/uuid"
	"github.com/khoahotran/personal-os/internal/domain/project"
	"github.com/khoahotran/personal-os/internal/domain/tag"
	"github.com/khoahotran/personal-os/pkg/logger"
	"go.uber.org/zap"
)

type GetProjectUseCase struct {
	projectRepo project.Repository
	tagRepo     tag.Repository
	logger      logger.Logger
}

func NewGetProjectUseCase(pRepo project.Repository, tRepo tag.Repository, log logger.Logger) *GetProjectUseCase {
	return &GetProjectUseCase{projectRepo: pRepo, tagRepo: tRepo, logger: log}
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
	tags, err := uc.tagRepo.GetTagsForResource(ctx, p.ID, "project")
	if err != nil {
		uc.logger.Warn("Failed to get tags for project", zap.String("project_id", p.ID.String()), zap.Error(err))
	}
	return &GetProjectOutput{Project: p, Tags: tags}, nil
}

type GetPublicProjectUseCase struct {
	projectRepo project.Repository
	tagRepo     tag.Repository
	logger      logger.Logger
}

func NewGetPublicProjectUseCase(pRepo project.Repository, tRepo tag.Repository, log logger.Logger) *GetPublicProjectUseCase {
	return &GetPublicProjectUseCase{projectRepo: pRepo, tagRepo: tRepo, logger: log}
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
		uc.logger.Warn("Failed to get tags for public project", zap.String("project_id", p.ID.String()), zap.Error(err))
	}
	return &GetPublicProjectOutput{Project: p, Tags: tags}, nil
}
