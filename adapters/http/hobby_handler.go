package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	hobbyUC "github.com/khoahotran/personal-os/internal/application/usecase/hobby"
)

type HobbyHandler struct {
	useCase *hobbyUC.HobbyUseCase
}

func NewHobbyHandler(uc *hobbyUC.HobbyUseCase) *HobbyHandler {
	return &HobbyHandler{useCase: uc}
}

func (h *HobbyHandler) CreateHobbyItem(c *gin.Context) {
	ownerID, ok := GetOwnerIDFromGinContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "owner information not found"})
		return
	}

	var req CreateOrUpdateHobbyItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request data", "details": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create hobby item failed", "details": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, ToHobbyItemDTO(item))
}

func (h *HobbyHandler) UpdateHobbyItem(c *gin.Context) {
	ownerID, ok := GetOwnerIDFromGinContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "owner information not found"})
		return
	}
	itemID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid hobby ID"})
		return
	}

	var req CreateOrUpdateHobbyItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request data"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update hobby item failed", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ToHobbyItemDTO(item))
}

func (h *HobbyHandler) DeleteHobbyItem(c *gin.Context) {
	ownerID, ok := GetOwnerIDFromGinContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "owner information not found"})
		return
	}
	itemID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid hobby ID"})
		return
	}

	if err := h.useCase.DeleteHobbyItem(c.Request.Context(), itemID, ownerID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete hobby item failed", "details": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *HobbyHandler) GetHobbyItem(c *gin.Context) {
	ownerID, ok := GetOwnerIDFromGinContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "owner information not found"})
		return
	}
	itemID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid hobby ID"})
		return
	}

	item, err := h.useCase.GetHobbyItem(c.Request.Context(), itemID, ownerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "hobby item not found"})
		return
	}
	c.JSON(http.StatusOK, ToHobbyItemDTO(item))
}

func (h *HobbyHandler) ListHobbyItems(c *gin.Context) {
	ownerID, ok := GetOwnerIDFromGinContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "owner information not found"})
		return
	}
	category := c.Query("category")
	if category == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "'category' query param is required"})
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	items, err := h.useCase.ListHobbyItems(c.Request.Context(), ownerID, category, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "get hobby items failed"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "'category' query param is required"})
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "30"))

	items, err := h.useCase.ListPublicHobbyItems(c.Request.Context(), category, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "get public hobby items failed"})
		return
	}
	dtos := make([]HobbyItemDTO, len(items))
	for i, item := range items {
		dtos[i] = ToHobbyItemDTO(item)
	}
	c.JSON(http.StatusOK, dtos)
}
