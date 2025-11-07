package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	searchUC "github.com/khoahotran/personal-os/internal/application/usecase/search"
	"github.com/khoahotran/personal-os/pkg/apperror"
	"github.com/khoahotran/personal-os/pkg/logger"
)

type SearchHandler struct {
	searchUseCase *searchUC.SearchUseCase
	logger        logger.Logger
}

func NewSearchHandler(uc *searchUC.SearchUseCase, log logger.Logger) *SearchHandler {
	return &SearchHandler{
		searchUseCase: uc,
		logger:        log,
	}
}

func (h *SearchHandler) handleSearch(c *gin.Context, isPublic bool) {

	query := c.Query("q")
	if query == "" {
		c.Error(apperror.NewInvalidInput("'q' query param is required", nil))
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	var ownerID uuid.UUID
	if !isPublic {
		var ok bool
		ownerID, ok = GetOwnerIDFromGinContext(c)
		if !ok {
			c.Error(apperror.NewPermissionDenied("ownerID not found in context"))
			return
		}
	}

	input := searchUC.SearchInput{
		Query:    query,
		OwnerID:  ownerID,
		IsPublic: isPublic,
		Limit:    limit,
	}
	output, err := h.searchUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		c.Error(err)
		return
	}

	dtos := make([]SearchResultDTO, len(output.Results))
	for i, res := range output.Results {
		dtos[i] = ToSearchResultDTO(res)
	}
	c.JSON(http.StatusOK, dtos)
}

func (h *SearchHandler) SearchPublic(c *gin.Context) {
	h.handleSearch(c, true)
}

func (h *SearchHandler) SearchPrivate(c *gin.Context) {
	h.handleSearch(c, false)
}
