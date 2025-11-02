package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	mediaUC "github.com/khoahotran/personal-os/internal/application/usecase/media"
)

type MediaHandler struct {
	uploadMediaUC *mediaUC.UploadMediaUseCase
	listPublicUC  *mediaUC.ListPublicMediaUseCase
	updateMediaUC *mediaUC.UpdateMediaUseCase
	deleteMediaUC *mediaUC.DeleteMediaUseCase
}

func NewMediaHandler(
	uploadUC *mediaUC.UploadMediaUseCase,
	listPublicUC *mediaUC.ListPublicMediaUseCase,
	updateUC *mediaUC.UpdateMediaUseCase,
	deleteUC *mediaUC.DeleteMediaUseCase,
) *MediaHandler {
	return &MediaHandler{
		uploadMediaUC: uploadUC,
		listPublicUC:  listPublicUC,
		updateMediaUC: updateUC,
		deleteMediaUC: deleteUC,
	}
}

func (h *MediaHandler) UploadMedia(c *gin.Context) {
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
	if err != nil { /* ... */
	}
	defer file.Close()

	dataJSON := c.PostForm("data")
	var reqData struct {
		Metadata map[string]any `json:"metadata"`
		IsPublic bool           `json:"is_public"`
	}
	if dataJSON != "" {
		if err := json.Unmarshal([]byte(dataJSON), &reqData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "field 'data' must be valid JSON"})
			return
		}
	}

	if reqData.Metadata == nil {
		reqData.Metadata = make(map[string]any)
	}
	reqData.Metadata["original_filename"] = fileHeader.Filename

	input := mediaUC.UploadMediaInput{
		OwnerID:  ownerID,
		File:     file,
		Metadata: reqData.Metadata,
		IsPublic: reqData.IsPublic,
		Provider: "cloudinary",
	}

	output, err := h.uploadMediaUC.Execute(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "upload media failed", "details": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Upload media success, processing...", "media_id": output.MediaID})
}

func (h *MediaHandler) UpdateMedia(c *gin.Context) {
	ownerID, ok := GetOwnerIDFromGinContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "owner information not found"})
		return
	}
	mediaID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid media ID"})
		return
	}

	var req UpdateMediaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request data"})
		return
	}

	input := mediaUC.UpdateMediaInput{
		OwnerID:  ownerID,
		MediaID:  mediaID,
		Metadata: req.Metadata,
		IsPublic: req.IsPublic,
	}

	if err := h.updateMediaUC.Execute(c.Request.Context(), input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update media failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "update media successfully"})
}

func (h *MediaHandler) DeleteMedia(c *gin.Context) {
	ownerID, ok := GetOwnerIDFromGinContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "owner information not found"})
		return
	}
	mediaID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid media ID"})
		return
	}

	input := mediaUC.DeleteMediaInput{OwnerID: ownerID, MediaID: mediaID}
	if err := h.deleteMediaUC.Execute(c.Request.Context(), input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete media failed"})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *MediaHandler) ListPublicMedia(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "30"))

	input := mediaUC.ListPublicMediaInput{Limit: limit, Offset: (page - 1) * limit}
	output, err := h.listPublicUC.Execute(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "get public media failed"})
		return
	}

	dtos := make([]MediaDTO, len(output.Medias))
	for i, m := range output.Medias {
		dtos[i] = ToMediaDTO(m)
	}
	c.JSON(http.StatusOK, dtos)
}
