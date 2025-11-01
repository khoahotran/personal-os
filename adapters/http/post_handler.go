package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	postUC "github.com/khoahotran/personal-os/internal/application/usecase/post"
	"github.com/khoahotran/personal-os/internal/domain/post"
)

type PostHandler struct {
	createPostUseCase      *postUC.CreatePostUseCase
	listPostsUseCase       *postUC.ListPostsUseCase
	listPublicPostsUseCase *postUC.ListPublicPostsUseCase
	updatePostUseCase      *postUC.UpdatePostUseCase
	deletePostUseCase      *postUC.DeletePostUseCase
	getPostUseCase         *postUC.GetPostUseCase
	getPublicPostUseCase   *postUC.GetPublicPostUseCase
}

func NewPostHandler(
	createUC *postUC.CreatePostUseCase,
	listUC *postUC.ListPostsUseCase,
	listPublicUC *postUC.ListPublicPostsUseCase,
	updateUC *postUC.UpdatePostUseCase,
	deleteUC *postUC.DeletePostUseCase,
	getUC *postUC.GetPostUseCase,
	getPublicUC *postUC.GetPublicPostUseCase,
) *PostHandler {
	return &PostHandler{
		createPostUseCase:      createUC,
		listPostsUseCase:       listUC,
		listPublicPostsUseCase: listPublicUC,
		updatePostUseCase:      updateUC,
		deletePostUseCase:      deleteUC,
		getPostUseCase:         getUC,
		getPublicPostUseCase:   getPublicUC,
	}
}

func (h *PostHandler) CreatePost(c *gin.Context) {

	ownerID, ok := GetOwnerIDFromGinContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "owner information not found"})
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "'file' is required"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "file cannot open"})
		return
	}
	defer file.Close()

	title := c.PostForm("title")
	content := c.PostForm("content")
	slug := c.PostForm("slug")
	status := c.PostForm("status") 
	tagsJSON := c.PostForm("tags") 

	if title == "" || status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "'title' and 'status' is required"})
		return
	}

	var tagNames []string
	if tagsJSON != "" {
		if err := json.Unmarshal([]byte(tagsJSON), &tagNames); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "'tags' field is invalid JSON array"})
			return
		}
	}

	var reqStatus post.PostStatus
	switch status {
	case "public":
		reqStatus = post.StatusPublic
	case "private":
		reqStatus = post.StatusPrivate
	default:
		reqStatus = post.StatusDraft
	}

	input := postUC.CreatePostInput{
		OwnerID:         ownerID,
		Title:           title,
		Content:         content,
		Slug:            slug,
		RequestedStatus: reqStatus, 
		TagNames:        tagNames,
		File:            file,
		Metadata:        map[string]any{"original_filename": fileHeader.Filename, "requested_status": string(reqStatus)},
	}

	output, err := h.createPostUseCase.Execute(c.Request.Context(), input)
	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{"error": "create post failed", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "created post successfully, processing ...",
		"post_id": output.PostID,
		"slug":    output.Slug,
	})
}

func (h *PostHandler) ListPosts(c *gin.Context) {

	ownerID, ok := GetOwnerIDFromGinContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "owner information not found"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "get post list failed"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "get public post list failed"})
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "owner information not found"})
		return
	}

	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post ID"})
		return
	}

	var req CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request data", "details": err.Error()})
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
		if errors.Is(err, errors.New("post not found")) {
			c.JSON(http.StatusNotFound, gin.H{"error": "post not found or no permission"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update post failed", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ToPostSummaryDTO(output.Post))
}

func (h *PostHandler) DeletePost(c *gin.Context) {

	ownerID, ok := GetOwnerIDFromGinContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "owner information not found"})
		return
	}

	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post ID"})
		return
	}

	input := postUC.DeletePostInput{
		PostID:  postID,
		OwnerID: ownerID,
	}

	if err := h.deletePostUseCase.Execute(c.Request.Context(), input); err != nil {
		if errors.Is(err, errors.New("post not found")) {
			c.JSON(http.StatusNotFound, gin.H{"error": "post not found or no permission"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete post failed", "details": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *PostHandler) GetPost(c *gin.Context) {

	ownerID, ok := GetOwnerIDFromGinContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "owner information not found"})
		return
	}

	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post ID"})
		return
	}

	input := postUC.GetPostInput{
		PostID:  postID,
		OwnerID: ownerID,
	}
	output, err := h.getPostUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		if errors.Is(err, errors.New("post not found")) {
			c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "get post failed"})
		return
	}

	c.JSON(http.StatusOK, ToPostDTO(output.Post, output.Tags))
}

func (h *PostHandler) GetPublicPost(c *gin.Context) {

	slug := c.Param("slug")

	input := postUC.GetPublicPostInput{Slug: slug}
	output, err := h.getPublicPostUseCase.Execute(c.Request.Context(), input)

	if err != nil {

		if errors.Is(err, post.ErrPostNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "get public post failed"})
		return
	}

	c.JSON(http.StatusOK, ToPostDTO(output.Post, output.Tags))
}
