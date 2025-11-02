package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/khoahotran/personal-os/internal/application/usecase/auth"
	"github.com/khoahotran/personal-os/pkg/apperror"
	"github.com/khoahotran/personal-os/pkg/logger"
)

type AuthHandler struct {
	loginUseCase *auth.LoginUseCase
	logger       logger.Logger
}

func NewAuthHandler(loginUC *auth.LoginUseCase, log logger.Logger) *AuthHandler {
	return &AuthHandler{
		loginUseCase: loginUC,
		logger:       log,
	}
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := apperror.NewInvalidInput("invalid JSON body", err)
		c.Error(appErr)
		return
	}

	input := auth.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	}

	output, err := h.loginUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": output.AccessToken,
	})
}
