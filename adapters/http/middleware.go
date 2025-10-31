package http

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/khoahotran/personal-os/pkg/auth"
)

const (
	GinContextKeyOwnerID = "ownerID"
)

func AuthMiddleware(jwtSvc *auth.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			return
		}

		claims, err := jwtSvc.ValidateToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
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
