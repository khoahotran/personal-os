package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	chatUC "github.com/khoahotran/personal-os/internal/application/usecase/chat"
	"github.com/khoahotran/personal-os/pkg/apperror"
	"github.com/khoahotran/personal-os/pkg/logger"
)

type ChatHandler struct {
	chatUseCase *chatUC.ChatUseCase
	logger      logger.Logger
}

func NewChatHandler(uc *chatUC.ChatUseCase, log logger.Logger) *ChatHandler {
	return &ChatHandler{
		chatUseCase: uc,
		logger:      log,
	}
}

func (h *ChatHandler) Chat(c *gin.Context) {

	ownerID, ok := GetOwnerIDFromGinContext(c)
	if !ok {
		c.Error(apperror.NewPermissionDenied("ownerID not found in context"))
		return
	}

	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperror.NewInvalidInput("invalid JSON body", err))
		return
	}

	input := chatUC.ChatInput{
		Query:   req.Query,
		OwnerID: ownerID,
		Limit:   req.Limit,
	}

	output, err := h.chatUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		c.Error(err)
		return
	}

	sourcesDTO := make([]PostSummaryDTO, len(output.Sources))
	for i, p := range output.Sources {
		sourcesDTO[i] = ToPostSummaryDTO(p)
	}

	c.JSON(http.StatusOK, ChatResponse{
		Response: output.Response,
		Sources:  sourcesDTO,
	})
}
