package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	postUC "github.com/khoahotran/personal-os/internal/application/usecase/post"
	"github.com/khoahotran/personal-os/internal/domain/post"
	"github.com/khoahotran/personal-os/pkg/apperror"
	"github.com/khoahotran/personal-os/pkg/logger"
)

type PostHandler struct {
	createPostUseCase      *postUC.CreatePostUseCase
	listPostsUseCase       *postUC.ListPostsUseCase
	listPublicPostsUseCase *postUC.ListPublicPostsUseCase
	updatePostUseCase      *postUC.UpdatePostUseCase
	deletePostUseCase      *postUC.DeletePostUseCase
	getPostUseCase         *postUC.GetPostUseCase
	getPublicPostUseCase   *postUC.GetPublicPostUseCase
	logger                 logger.Logger
}

func NewPostHandler(
	createUC *postUC.CreatePostUseCase,
	listUC *postUC.ListPostsUseCase,
	listPublicUC *postUC.ListPublicPostsUseCase,
	updateUC *postUC.UpdatePostUseCase,
	deleteUC *postUC.DeletePostUseCase,
	getUC *postUC.GetPostUseCase,
	getPublicUC *postUC.GetPublicPostUseCase,
	log logger.Logger,
) *PostHandler {
	return &PostHandler{
		createPostUseCase:      createUC,
		listPostsUseCase:       listUC,
		listPublicPostsUseCase: listPublicUC,
		updatePostUseCase:      updateUC,
		deletePostUseCase:      deleteUC,
		getPostUseCase:         getUC,
		getPublicPostUseCase:   getPublicUC,
		logger:                 log,
	}
}

func (h *PostHandler) CreatePost(c *gin.Context) {

	ownerID, ok := GetOwnerIDFromGinContext(c)
	if !ok {
		c.Error(apperror.NewPermissionDenied("ownerID not found in context"))
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.Error(apperror.NewInvalidInput("'file' is required", err))
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.Error(apperror.NewInternal("failed to open file", err))
		return
	}
	defer file.Close()

	dataJSON := c.PostForm("data")
	if dataJSON == "" {
		c.Error(apperror.NewInvalidInput("'data' (JSON string) is required", nil))
		return
	}
	var reqData struct {
		Title   string   `json:"title"`
		Content string   `json:"content"`
		Slug    string   `json:"slug"`
		Status  string   `json:"status"`
		Tags    []string `json:"tags"`
	}
	if err := json.Unmarshal([]byte(dataJSON), &reqData); err != nil {
		c.Error(apperror.NewInvalidInput("'data' field is not valid JSON", err))
		return
	}
	if reqData.Title == "" || reqData.Status == "" {
		c.Error(apperror.NewInvalidInput("'title' and 'status' are required in data", nil))
		return
	}

	var reqStatus post.PostStatus
	switch reqData.Status {
	case "public":
		reqStatus = post.StatusPublic
	case "private":
		reqStatus = post.StatusPrivate
	default:
		reqStatus = post.StatusDraft
	}

	input := postUC.CreatePostInput{
		OwnerID:         ownerID,
		Title:           reqData.Title,
		Content:         reqData.Content,
		Slug:            reqData.Slug,
		RequestedStatus: reqStatus,
		TagNames:        reqData.Tags,
		File:            file,
		Metadata:        map[string]any{"original_filename": fileHeader.Filename, "requested_status": string(reqStatus)},
	}

	output, err := h.createPostUseCase.Execute(c.Request.Context(), input)
	if err != nil {

		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "created post successfully, processing ...", "post_id": output.PostID, "slug": output.Slug})
}

func (h *PostHandler) ListPosts(c *gin.Context) {

	ownerID, ok := GetOwnerIDFromGinContext(c)
	if !ok {
		c.Error(apperror.NewPermissionDenied("ownerID not found in context"))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	input := postUC.ListPostsInput{
		OwnerID: ownerID,
		Page:    page,
		Limit:   limit,
	}
	output, err := h.listPostsUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		c.Error(err)
		return
	}

	dtos := make([]PostSummaryDTO, len(output.Posts))
	for i, p := range output.Posts {
		dtos[i] = ToPostSummaryDTO(p)
	}
	c.JSON(http.StatusOK, dtos)
}

func (h *PostHandler) ListPublicPosts(c *gin.Context) {

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	input := postUC.ListPublicPostsInput{
		Page:  page,
		Limit: limit,
	}
	output, err := h.listPublicPostsUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		c.Error(err)
		return
	}

	dtos := make([]PostSummaryDTO, len(output.Posts))
	for i, p := range output.Posts {
		dtos[i] = ToPostSummaryDTO(p)
	}
	c.JSON(http.StatusOK, dtos)
}

func (h *PostHandler) UpdatePost(c *gin.Context) {

	ownerID, ok := GetOwnerIDFromGinContext(c)
	if !ok {
		c.Error(apperror.NewPermissionDenied("ownerID not found in context"))
		return
	}

	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.NewInvalidInput("invalid post ID", err))
		return
	}

	var req UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperror.NewInvalidInput("invalid request data", err))
		return
	}

	input := postUC.UpdatePostInput{
		PostID:  postID,
		OwnerID: ownerID,
		Title:   req.Title,
		Content: req.Content,
		Slug:    req.Slug,
		Status:  req.ToDomainPostStatus(),
		Tags:    req.Tags,
	}

	output, err := h.updatePostUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ToPostSummaryDTO(output.Post))
}

func (h *PostHandler) DeletePost(c *gin.Context) {

	ownerID, ok := GetOwnerIDFromGinContext(c)
	if !ok {
		c.Error(apperror.NewPermissionDenied("ownerID not found in context"))
		return
	}

	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.NewInvalidInput("invalid post ID", err))
		return
	}

	input := postUC.DeletePostInput{
		PostID:  postID,
		OwnerID: ownerID,
	}

	if err := h.deletePostUseCase.Execute(c.Request.Context(), input); err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *PostHandler) GetPost(c *gin.Context) {

	ownerID, ok := GetOwnerIDFromGinContext(c)
	if !ok {
		c.Error(apperror.NewPermissionDenied("ownerID not found in context"))
		return
	}

	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.NewInvalidInput("invalid post ID", err))
		return
	}

	input := postUC.GetPostInput{
		PostID:  postID,
		OwnerID: ownerID,
	}
	output, err := h.getPostUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ToPostDTO(output.Post, output.Tags))
}

func (h *PostHandler) GetPublicPost(c *gin.Context) {
	slug := c.Param("slug")

	input := postUC.GetPublicPostInput{Slug: slug}
	output, err := h.getPublicPostUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ToPostDTO(output.Post, output.Tags))
}
