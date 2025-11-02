package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	projectUC "github.com/khoahotran/personal-os/internal/application/usecase/project"
	"github.com/khoahotran/personal-os/internal/domain/project"
	"github.com/khoahotran/personal-os/pkg/apperror"
	"github.com/khoahotran/personal-os/pkg/logger"
)

type ProjectHandler struct {
	createProjectUseCase      *projectUC.CreateProjectUseCase
	listProjectsUseCase       *projectUC.ListProjectsUseCase
	listPublicProjectsUseCase *projectUC.ListPublicProjectsUseCase
	getProjectUseCase         *projectUC.GetProjectUseCase
	getPublicProjectUseCase   *projectUC.GetPublicProjectUseCase
	updateProjectUseCase      *projectUC.UpdateProjectUseCase
	deleteProjectUseCase      *projectUC.DeleteProjectUseCase
	logger                    logger.Logger
}

func NewProjectHandler(
	createUC *projectUC.CreateProjectUseCase,
	listUC *projectUC.ListProjectsUseCase,
	listPublicUC *projectUC.ListPublicProjectsUseCase,
	getUC *projectUC.GetProjectUseCase,
	getPublicUC *projectUC.GetPublicProjectUseCase,
	updateUC *projectUC.UpdateProjectUseCase,
	deleteUC *projectUC.DeleteProjectUseCase,
	log logger.Logger,
) *ProjectHandler {
	return &ProjectHandler{
		createProjectUseCase:      createUC,
		listProjectsUseCase:       listUC,
		listPublicProjectsUseCase: listPublicUC,
		getProjectUseCase:         getUC,
		getPublicProjectUseCase:   getPublicUC,
		updateProjectUseCase:      updateUC,
		deleteProjectUseCase:      deleteUC,
		logger:                    log,
	}
}

func (h *ProjectHandler) CreateProject(c *gin.Context) {
	ownerID, ok := GetOwnerIDFromGinContext(c)
	if !ok {
		c.Error(apperror.NewPermissionDenied("ownerID not found in context"))
		return
	}
	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperror.NewInvalidInput("invalid request data", err))
		return
	}

	input := projectUC.CreateProjectInput{
		OwnerID:       ownerID,
		Title:         req.Title,
		Slug:          req.Slug,
		Description:   req.Description,
		Stack:         req.Stack,
		RepositoryURL: req.RepositoryURL,
		LiveURL:       req.LiveURL,
		IsPublic:      req.IsPublic,
		TagNames:      req.TagNames,
		Media:         []project.ProjectMedia{},
	}

	output, err := h.createProjectUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"project_id": output.ProjectID, "slug": output.Slug})
}

func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	ownerID, ok := GetOwnerIDFromGinContext(c)
	if !ok {
		c.Error(apperror.NewPermissionDenied("ownerID not found in context"))
		return
	}
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.NewInvalidInput("invalid project ID", err))
		return
	}
	var req UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperror.NewInvalidInput("invalid request data", err))
		return
	}

	input := projectUC.UpdateProjectInput{
		ProjectID:     projectID,
		OwnerID:       ownerID,
		Title:         req.Title,
		Slug:          req.Slug,
		Description:   req.Description,
		Stack:         req.Stack,
		RepositoryURL: req.RepositoryURL,
		LiveURL:       req.LiveURL,
		IsPublic:      req.IsPublic,
		TagNames:      req.TagNames,
		Media:         []project.ProjectMedia{},
	}

	output, err := h.updateProjectUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ToProjectSummaryDTO(output.Project))
}

func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	ownerID, ok := GetOwnerIDFromGinContext(c)
	if !ok {
		c.Error(apperror.NewPermissionDenied("ownerID not found in context"))
		return
	}
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.NewInvalidInput("invalid project ID", err))
		return
	}

	input := projectUC.DeleteProjectInput{ProjectID: projectID, OwnerID: ownerID}
	if err := h.deleteProjectUseCase.Execute(c.Request.Context(), input); err != nil {
		c.Error(err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ProjectHandler) GetProject(c *gin.Context) {
	ownerID, ok := GetOwnerIDFromGinContext(c)
	if !ok {
		c.Error(apperror.NewPermissionDenied("ownerID not found in context"))
		return
	}
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.NewInvalidInput("invalid project ID", err))
		return
	}
	input := projectUC.GetProjectInput{ProjectID: projectID, OwnerID: ownerID}
	output, err := h.getProjectUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ToProjectDTO(output.Project, output.Tags))
}

func (h *ProjectHandler) ListProjects(c *gin.Context) {
	ownerID, ok := GetOwnerIDFromGinContext(c)
	if !ok {
		c.Error(apperror.NewPermissionDenied("ownerID not found in context"))
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	input := projectUC.ListProjectsInput{OwnerID: ownerID, Page: page, Limit: limit}
	output, err := h.listProjectsUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		c.Error(err)
		return
	}
	dtos := make([]ProjectSummaryDTO, len(output.Projects))
	for i, p := range output.Projects {
		dtos[i] = ToProjectSummaryDTO(p)
	}
	c.JSON(http.StatusOK, dtos)
}

func (h *ProjectHandler) GetPublicProject(c *gin.Context) {
	slug := c.Param("slug")
	input := projectUC.GetPublicProjectInput{Slug: slug}
	output, err := h.getPublicProjectUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ToProjectDTO(output.Project, output.Tags))
}

func (h *ProjectHandler) ListPublicProjects(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	input := projectUC.ListPublicProjectsInput{Page: page, Limit: limit}
	output, err := h.listPublicProjectsUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		c.Error(err)
		return
	}
	dtos := make([]ProjectSummaryDTO, len(output.Projects))
	for i, p := range output.Projects {
		dtos[i] = ToProjectSummaryDTO(p)
	}
	c.JSON(http.StatusOK, dtos)
}
