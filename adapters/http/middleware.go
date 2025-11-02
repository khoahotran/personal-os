package http

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/khoahotran/personal-os/pkg/apperror"
	"github.com/khoahotran/personal-os/pkg/auth"
	"github.com/khoahotran/personal-os/pkg/logger"
	"go.uber.org/zap"
)

const (
	GinContextKeyOwnerID = "ownerID"
)

func AuthMiddleware(jwtSvc *auth.JWTService, log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			err := apperror.NewAppError(apperror.ErrUnauthorized, "Authorization header is required", "no auth header", nil)
			c.Error(err)
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			err := apperror.NewAppError(apperror.ErrUnauthorized, "Invalid token format", "bearer prefix missing", nil)
			c.Error(err)
			c.Abort()
			return
		}

		claims, err := jwtSvc.ValidateToken(tokenString)
		if err != nil {
			appErr := apperror.NewAppError(apperror.ErrUnauthorized, "Invalid or expired token", "token validation failed", err)
			log.Warn("Token validation failed", zap.Error(err))
			c.Error(appErr)
			c.Abort()
			return
		}

		c.Set(GinContextKeyOwnerID, claims.OwnerID)
		c.Next()
	}
}

func GetOwnerIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	ownerID, ok := ctx.Value(GinContextKeyOwnerID).(uuid.UUID)
	return ownerID, ok
}

func GetOwnerIDFromGinContext(c *gin.Context) (uuid.UUID, bool) {
	ownerID, ok := c.Get(GinContextKeyOwnerID)
	if !ok {
		return uuid.Nil, false
	}
	ownerIDUUID, ok := ownerID.(uuid.UUID)
	if !ok {
		return uuid.Nil, false
	}
	return ownerIDUUID, true
}

func ErrorMiddleware(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		err := c.Errors.Last()
		if err == nil {
			return
		}

		originalErr := err.Err

		var appErr *apperror.AppError

		if errors.As(originalErr, &appErr) {
			statusCode := apperror.ToHTTPStatus(appErr)

			if statusCode >= 500 {

				log.Error(appErr.Details, appErr.Err, zap.String("path", c.Request.URL.Path))
			} else {

				log.Warn(appErr.Details, zap.String("path", c.Request.URL.Path), zap.Error(appErr.Err))
			}

			c.JSON(statusCode, appErr.ToJSON())
			return
		}

		log.Error("Unhandled internal error", originalErr, zap.String("path", c.Request.URL.Path))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal server error",
			"message": "An unexpected error occurred",
		})
	}
}
