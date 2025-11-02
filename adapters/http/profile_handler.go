package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	profileUC "github.com/khoahotran/personal-os/internal/application/usecase/profile"
	"github.com/khoahotran/personal-os/pkg/apperror"
	"github.com/khoahotran/personal-os/pkg/logger"
)

type ProfileHandler struct {
	profileUseCase *profileUC.ProfileUseCase
	logger         logger.Logger
}

func NewProfileHandler(uc *profileUC.ProfileUseCase, log logger.Logger) *ProfileHandler {
	return &ProfileHandler{
		profileUseCase: uc,
		logger:         log,
	}
}

func (h *ProfileHandler) GetProfile(c *gin.Context) {
	ownerID, ok := GetOwnerIDFromGinContext(c)
	if !ok {
		c.Error(apperror.NewPermissionDenied("ownerID not found in context"))
		return
	}

	input := profileUC.GetProfileInput{OwnerID: ownerID}
	output, err := h.profileUseCase.ExecuteGetProfile(c.Request.Context(), input)
	if err != nil {
		c.Error(err)
		return
	}

	responseDTO := ToProfileDTO(output.Profile)
	c.JSON(http.StatusOK, responseDTO)
}

func (h *ProfileHandler) UpdateProfile(c *gin.Context) {
	ownerID, ok := GetOwnerIDFromGinContext(c)
	if !ok {
		c.Error(apperror.NewPermissionDenied("ownerID not found in context"))
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := apperror.NewInvalidInput("invalid JSON body for profile update", err)
		c.Error(appErr)
		return
	}

	input := profileUC.UpdateProfileInput{
		OwnerID:        ownerID,
		Bio:            req.Bio,
		CareerTimeline: req.ToDomainMilestones(),
		ThemeSettings:  req.ThemeSettings,
	}
	output, err := h.profileUseCase.ExecuteUpdateProfile(c.Request.Context(), input)
	if err != nil {
		c.Error(err)
		return
	}

	responseDTO := ToProfileDTO(output.Profile)
	c.JSON(http.StatusOK, responseDTO)
}
