package http

import (
	"github.com/gin-gonic/gin"
	postUC "github.com/khoahotran/personal-os/internal/application/usecase/post"
	"github.com/khoahotran/personal-os/pkg/apperror"
	"github.com/khoahotran/personal-os/pkg/logger"
)

type RSSHandler struct {
	rssUseCase *postUC.RSSUseCase
	logger     logger.Logger
}

func NewRSSHandler(uc *postUC.RSSUseCase, log logger.Logger) *RSSHandler {
	return &RSSHandler{
		rssUseCase: uc,
		logger:     log,
	}
}

func (h *RSSHandler) GenerateRSS(c *gin.Context) {

	feed, err := h.rssUseCase.Execute(c.Request.Context())
	if err != nil {
		c.Error(apperror.NewInternal("failed to generate RSS feed", err))
		return
	}

	c.Header("Content-Type", "application/xml; charset=utf-8")

	if err := feed.WriteRss(c.Writer); err != nil {

		h.logger.Error("Failed to write RSS feed to response", err)
	}
}
