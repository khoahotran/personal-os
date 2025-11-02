package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	hobbyUC "github.com/khoahotran/personal-os/internal/application/usecase/hobby"
	"github.com/khoahotran/personal-os/pkg/apperror"
	"github.com/khoahotran/personal-os/pkg/logger"
)

type HobbyHandler struct {
	useCase *hobbyUC.HobbyUseCase
	logger  logger.Logger
}

func NewHobbyHandler(uc *hobbyUC.HobbyUseCase, log logger.Logger) *HobbyHandler {
	return &HobbyHandler{useCase: uc, logger: log}
}

func (h *HobbyHandler) CreateHobbyItem(c *gin.Context) {
	ownerID, ok := GetOwnerIDFromGinContext(c)
	if !ok {
		c.Error(apperror.NewPermissionDenied("ownerID not found in context"))
		return
	}

	var req CreateOrUpdateHobbyItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperror.NewInvalidInput("invalid request data", err))
		return
	}

	input := hobbyUC.CreateHobbyItemInput{
		OwnerID:  ownerID,
		Category: req.Category,
		Title:    req.Title,
		Status:   req.Status,
		Rating:   req.Rating,
		Notes:    req.Notes,
		Metadata: req.Metadata,
		IsPublic: req.IsPublic,
	}

	item, err := h.useCase.CreateHobbyItem(c.Request.Context(), input)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, ToHobbyItemDTO(item))
}

func (h *HobbyHandler) UpdateHobbyItem(c *gin.Context) {
	ownerID, ok := GetOwnerIDFromGinContext(c)
	if !ok {
		c.Error(apperror.NewPermissionDenied("ownerID not found in context"))
		return
	}
	itemID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.NewInvalidInput("invalid item ID", err))
		return
	}

	var req CreateOrUpdateHobbyItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperror.NewInvalidInput("invalid request data", err))
		return
	}

	input := hobbyUC.UpdateHobbyItemInput{
		ItemID:   itemID,
		OwnerID:  ownerID,
		Category: req.Category,
		Title:    req.Title,
		Status:   req.Status,
		Rating:   req.Rating,
		Notes:    req.Notes,
		Metadata: req.Metadata,
		IsPublic: req.IsPublic,
	}

	item, err := h.useCase.UpdateHobbyItem(c.Request.Context(), input)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ToHobbyItemDTO(item))
}

func (h *HobbyHandler) DeleteHobbyItem(c *gin.Context) {
	ownerID, ok := GetOwnerIDFromGinContext(c)
	if !ok {
		c.Error(apperror.NewPermissionDenied("ownerID not found in context"))
		return
	}
	itemID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.NewInvalidInput("invalid item ID", err))
		return
	}

	if err := h.useCase.DeleteHobbyItem(c.Request.Context(), itemID, ownerID); err != nil {
		c.Error(err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *HobbyHandler) GetHobbyItem(c *gin.Context) {
	ownerID, ok := GetOwnerIDFromGinContext(c)
	if !ok {
		c.Error(apperror.NewPermissionDenied("ownerID not found in context"))
		return
	}
	itemID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.NewInvalidInput("invalid item ID", err))
		return
	}

	item, err := h.useCase.GetHobbyItem(c.Request.Context(), itemID, ownerID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ToHobbyItemDTO(item))
}

func (h *HobbyHandler) ListHobbyItems(c *gin.Context) {
	ownerID, ok := GetOwnerIDFromGinContext(c)
	if !ok {
		c.Error(apperror.NewPermissionDenied("ownerID not found in context"))
		return
	}
	category := c.Query("category")
	if category == "" {
		c.Error(apperror.NewInvalidInput("'category' query param is required", nil))
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	items, err := h.useCase.ListHobbyItems(c.Request.Context(), ownerID, category, page, limit)
	if err != nil {
		c.Error(err)
		return
	}
	dtos := make([]HobbyItemDTO, len(items))
	for i, item := range items {
		dtos[i] = ToHobbyItemDTO(item)
	}
	c.JSON(http.StatusOK, dtos)
}

func (h *HobbyHandler) ListPublicHobbyItems(c *gin.Context) {
	category := c.Query("category")
	if category == "" {
		c.Error(apperror.NewInvalidInput("'category' query param is required", nil))
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1D"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "30"))

	items, err := h.useCase.ListPublicHobbyItems(c.Request.Context(), category, page, limit)
	if err != nil {
		c.Error(err)
		return
	}
	dtos := make([]HobbyItemDTO, len(items))
	for i, item := range items {
		dtos[i] = ToHobbyItemDTO(item)
	}
	c.JSON(http.StatusOK, dtos)
}
