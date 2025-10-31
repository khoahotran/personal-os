package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	profileUC "github.com/khoahotran/personal-os/internal/application/usecase/profile"
)

type ProfileHandler struct {
	profileUseCase *profileUC.ProfileUseCase
}

func NewProfileHandler(uc *profileUC.ProfileUseCase) *ProfileHandler {
	return &ProfileHandler{
		profileUseCase: uc,
	}
}

func (h *ProfileHandler) GetProfile(c *gin.Context) {
	ownerID, ok := GetOwnerIDFromGinContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "owner information not found"})
		return
	}

	input := profileUC.GetProfileInput{OwnerID: ownerID}
	output, err := h.profileUseCase.ExecuteGetProfile(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "profile retrieval failed"})
		return
	}

	c.JSON(http.StatusOK, ToProfileDTO(output.Profile))
}

func (h *ProfileHandler) UpdateProfile(c *gin.Context) {
	ownerID, ok := GetOwnerIDFromGinContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "owner information not found"})
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request data", "details": err.Error()})
		return
	}

	input := profileUC.UpdateProfileInput{
		OwnerID:        ownerID,
		Bio:            req.Bio,
		CareerTimeline: req.ToDomainMilestones(),
	}
	output, err := h.profileUseCase.ExecuteUpdateProfile(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "profile update failed"})
		return
	}

	c.JSON(http.StatusOK, ToProfileDTO(output.Profile))
}
