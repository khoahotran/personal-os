package project

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/khoahotran/personal-os/internal/domain/project"
	"github.com/khoahotran/personal-os/internal/domain/tag"
	"github.com/khoahotran/personal-os/pkg/apperror"
	"github.com/khoahotran/personal-os/pkg/logger"
	"go.uber.org/zap"
)

type CreateProjectUseCase struct {
	projectRepo project.Repository
	tagRepo     tag.Repository
	logger      logger.Logger
}

func NewCreateProjectUseCase(pRepo project.Repository, tRepo tag.Repository, log logger.Logger) *CreateProjectUseCase {
	return &CreateProjectUseCase{
		projectRepo: pRepo,
		tagRepo:     tRepo,
		logger:      log,
	}
}

type CreateProjectInput struct {
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
type CreateProjectOutput struct {
	ProjectID uuid.UUID
	Slug      string
}

func (uc *CreateProjectUseCase) Execute(ctx context.Context, input CreateProjectInput) (*CreateProjectOutput, error) {
	if input.Slug == "" {
		input.Slug = strings.ToLower(strings.ReplaceAll(input.Title, " ", "-"))
	}
	now := time.Now().UTC()

	newProject := &project.Project{
		ID:            uuid.New(),
		OwnerID:       input.OwnerID,
		Slug:          input.Slug,
		Title:         input.Title,
		Description:   input.Description,
		Stack:         input.Stack,
		RepositoryURL: input.RepositoryURL,
		LiveURL:       input.LiveURL,
		Media:         input.Media,
		IsPublic:      input.IsPublic,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := newProject.Validate(); err != nil {
		return nil, apperror.NewInvalidInput("project validation failed", err)
	}

	tags, err := uc.tagRepo.FindOrCreateTags(ctx, input.TagNames)
	if err != nil {
		return nil, apperror.NewInternal("failed to process tags", err)
	}

	if err := uc.projectRepo.Save(ctx, newProject); err != nil {
		return nil, err
	}

	tagIDs := make([]uuid.UUID, len(tags))
	for i, t := range tags {
		tagIDs[i] = t.ID
	}
	if err = uc.tagRepo.SetTagsForResource(ctx, newProject.ID, "project", tagIDs); err != nil {
		uc.logger.Warn("Failed to set tags for new project", zap.String("project_id", newProject.ID.String()), zap.Error(err))
	}

	return &CreateProjectOutput{
		ProjectID: newProject.ID,
		Slug:      newProject.Slug,
	}, nil
}
