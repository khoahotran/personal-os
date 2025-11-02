package project

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/khoahotran/personal-os/internal/domain/project"
	"github.com/khoahotran/personal-os/internal/domain/tag"
)

type DeleteProjectUseCase struct {
	projectRepo project.Repository
	tagRepo     tag.Repository
}

func NewDeleteProjectUseCase(pRepo project.Repository, tRepo tag.Repository) *DeleteProjectUseCase {
	return &DeleteProjectUseCase{projectRepo: pRepo, tagRepo: tRepo}
}

type DeleteProjectInput struct {
	ProjectID uuid.UUID
	OwnerID   uuid.UUID
}

func (uc *DeleteProjectUseCase) Execute(ctx context.Context, input DeleteProjectInput) error {

	err := uc.tagRepo.SetTagsForResource(ctx, input.ProjectID, "project", []uuid.UUID{})
	if err != nil {
		return fmt.Errorf("delete tag link failed: %w", err)
	}

	err = uc.projectRepo.Delete(ctx, input.ProjectID, input.OwnerID)
	if err != nil {
		return fmt.Errorf("delete project failed: %w", err)
	}
	return nil
}
