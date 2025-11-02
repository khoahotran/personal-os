package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	projectUC "github.com/khoahotran/personal-os/internal/application/usecase/project"
	"github.com/khoahotran/personal-os/internal/domain/project"
)

type ProjectHandler struct {
	createProjectUseCase      *projectUC.CreateProjectUseCase
	listProjectsUseCase       *projectUC.ListProjectsUseCase
	listPublicProjectsUseCase *projectUC.ListPublicProjectsUseCase
	getProjectUseCase         *projectUC.GetProjectUseCase
	getPublicProjectUseCase   *projectUC.GetPublicProjectUseCase
	updateProjectUseCase      *projectUC.UpdateProjectUseCase
	deleteProjectUseCase      *projectUC.DeleteProjectUseCase
}

func NewProjectHandler(
	createUC *projectUC.CreateProjectUseCase,
	listUC *projectUC.ListProjectsUseCase,
	listPublicUC *projectUC.ListPublicProjectsUseCase,
	getUC *projectUC.GetProjectUseCase,
	getPublicUC *projectUC.GetPublicProjectUseCase,
	updateUC *projectUC.UpdateProjectUseCase,
	deleteUC *projectUC.DeleteProjectUseCase,
) *ProjectHandler {
	return &ProjectHandler{
		createProjectUseCase:      createUC,
		listProjectsUseCase:       listUC,
		listPublicProjectsUseCase: listPublicUC,
		getProjectUseCase:         getUC,
		getPublicProjectUseCase:   getPublicUC,
		updateProjectUseCase:      updateUC,
		deleteProjectUseCase:      deleteUC,
	}
}

func (h *ProjectHandler) CreateProject(c *gin.Context) {
	ownerID, ok := GetOwnerIDFromGinContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "owner information not found"})
		return
	}
	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request data", "details": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create project failed", "details": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"project_id": output.ProjectID, "slug": output.Slug})
}

func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	ownerID, ok := GetOwnerIDFromGinContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "owner information not found"})
		return
	}
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}
	var req UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request data", "details": err.Error()})
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
		if errors.Is(err, project.ErrProjectNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "project not found or no permission"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update project failed", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ToProjectSummaryDTO(output.Project))
}

func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	ownerID, ok := GetOwnerIDFromGinContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "owner information not found"})
		return
	}
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	input := projectUC.DeleteProjectInput{ProjectID: projectID, OwnerID: ownerID}
	if err := h.deleteProjectUseCase.Execute(c.Request.Context(), input); err != nil {
		if errors.Is(err, project.ErrProjectNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "project not found or no permission"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete project failed", "details": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ProjectHandler) GetProject(c *gin.Context) {
	ownerID, ok := GetOwnerIDFromGinContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "owner information not found"})
		return
	}
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	input := projectUC.GetProjectInput{ProjectID: projectID, OwnerID: ownerID}
	output, err := h.getProjectUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		if errors.Is(err, project.ErrProjectNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "get project failed"})
		return
	}
	c.JSON(http.StatusOK, ToProjectDTO(output.Project, output.Tags))
}

func (h *ProjectHandler) ListProjects(c *gin.Context) {
	ownerID, ok := GetOwnerIDFromGinContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "owner information not found"})
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	input := projectUC.ListProjectsInput{OwnerID: ownerID, Page: page, Limit: limit}
	output, err := h.listProjectsUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "get list project failed"})
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
		if errors.Is(err, project.ErrProjectNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "get public project failed"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "get list public project failed"})
		return
	}
	dtos := make([]ProjectSummaryDTO, len(output.Projects))
	for i, p := range output.Projects {
		dtos[i] = ToProjectSummaryDTO(p)
	}
	c.JSON(http.StatusOK, dtos)
}
