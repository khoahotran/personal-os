package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	mediaUC "github.com/khoahotran/personal-os/internal/application/usecase/media"
	"github.com/khoahotran/personal-os/pkg/apperror"
	"github.com/khoahotran/personal-os/pkg/logger"
)

type MediaHandler struct {
	uploadMediaUC *mediaUC.UploadMediaUseCase
	listPublicUC  *mediaUC.ListPublicMediaUseCase
	updateMediaUC *mediaUC.UpdateMediaUseCase
	deleteMediaUC *mediaUC.DeleteMediaUseCase
	logger        logger.Logger
}

func NewMediaHandler(
	uploadUC *mediaUC.UploadMediaUseCase,
	listPublicUC *mediaUC.ListPublicMediaUseCase,
	updateUC *mediaUC.UpdateMediaUseCase,
	deleteUC *mediaUC.DeleteMediaUseCase,
	log logger.Logger,
) *MediaHandler {
	return &MediaHandler{
		uploadMediaUC: uploadUC,
		listPublicUC:  listPublicUC,
		updateMediaUC: updateUC,
		deleteMediaUC: deleteUC,
		logger:        log,
	}
}

func (h *MediaHandler) UploadMedia(c *gin.Context) {
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
	var reqData struct {
		Metadata map[string]any `json:"metadata"`
		IsPublic bool           `json:"is_public"`
	}
	if dataJSON != "" {
		if err := json.Unmarshal([]byte(dataJSON), &reqData); err != nil {
			c.Error(apperror.NewInvalidInput("'data' field is not valid JSON", err))
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
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Upload media successfully, processing...", "media_id": output.MediaID})
}

func (h *MediaHandler) UpdateMedia(c *gin.Context) {
	ownerID, ok := GetOwnerIDFromGinContext(c)
	if !ok {
		c.Error(apperror.NewPermissionDenied("ownerID not found in context"))
		return
	}
	mediaID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.NewInvalidInput("invalid media ID", err))
		return
	}

	var req UpdateMediaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperror.NewInvalidInput("invalid request data", err))
		return
	}

	input := mediaUC.UpdateMediaInput{
		OwnerID:  ownerID,
		MediaID:  mediaID,
		Metadata: req.Metadata,
		IsPublic: req.IsPublic,
	}

	if err := h.updateMediaUC.Execute(c.Request.Context(), input); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Update media successfully"})
}

func (h *MediaHandler) DeleteMedia(c *gin.Context) {
	ownerID, ok := GetOwnerIDFromGinContext(c)
	if !ok {
		c.Error(apperror.NewPermissionDenied("ownerID not found in context"))
		return
	}
	mediaID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.NewInvalidInput("invalid media ID", err))
		return
	}

	input := mediaUC.DeleteMediaInput{OwnerID: ownerID, MediaID: mediaID}
	if err := h.deleteMediaUC.Execute(c.Request.Context(), input); err != nil {
		c.Error(err)
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
		c.Error(err)
		return
	}

	dtos := make([]MediaDTO, len(output.Medias))
	for i, m := range output.Medias {
		dtos[i] = ToMediaDTO(m)
	}
	c.JSON(http.StatusOK, dtos)
}
