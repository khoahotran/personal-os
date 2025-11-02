package project

import (
	"context"

	"github.com/google/uuid"
	"github.com/khoahotran/personal-os/internal/domain/project"
	"github.com/khoahotran/personal-os/internal/domain/tag"
	"github.com/khoahotran/personal-os/pkg/apperror"
	"github.com/khoahotran/personal-os/pkg/logger"
)

type DeleteProjectUseCase struct {
	projectRepo project.Repository
	tagRepo     tag.Repository
	logger      logger.Logger
}

func NewDeleteProjectUseCase(pRepo project.Repository, tRepo tag.Repository, log logger.Logger) *DeleteProjectUseCase {
	return &DeleteProjectUseCase{projectRepo: pRepo, tagRepo: tRepo, logger: log}
}

type DeleteProjectInput struct {
	ProjectID uuid.UUID
	OwnerID   uuid.UUID
}

func (uc *DeleteProjectUseCase) Execute(ctx context.Context, input DeleteProjectInput) error {
	err := uc.tagRepo.SetTagsForResource(ctx, input.ProjectID, "project", []uuid.UUID{})
	if err != nil {
		return apperror.NewInternal("failed to delete tag relations", err)
	}
	err = uc.projectRepo.Delete(ctx, input.ProjectID, input.OwnerID)
	if err != nil {
		return err
	}
	return nil
}
