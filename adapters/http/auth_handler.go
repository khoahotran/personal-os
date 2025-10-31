package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/khoahotran/personal-os/internal/application/usecase/auth"
)

type AuthHandler struct {
	loginUseCase *auth.LoginUseCase
}

func NewAuthHandler(loginUC *auth.LoginUseCase) *AuthHandler {
	return &AuthHandler{
		loginUseCase: loginUC,
	}
}

type loginRequest struct {
	Email string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input", "details": err.Error()})
		return
	}

	input := auth.LoginInput{
		Email: req.Email,
		Password: req.Password,
	}

	output, err := h.loginUseCase.Execute(c.Request.Context(), input)
	if err != nil {

		if errors.Is(err, auth.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Email or password is incorrect"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": output.AccessToken,
	})
}
