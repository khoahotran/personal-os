package auth

import (
	"context"
	"errors"

	"github.com/khoahotran/personal-os/internal/domain/user"
	"github.com/khoahotran/personal-os/pkg/apperror"
	"github.com/khoahotran/personal-os/pkg/auth"
	"github.com/khoahotran/personal-os/pkg/logger"
	"go.uber.org/zap"
)

var (
	ErrInvalidCredentials = errors.New("email or password is incorrect")
)

type LoginUseCase struct {
	userRepo user.Repository
	jwtSvc   *auth.JWTService
	logger   logger.Logger
}

func NewLoginUseCase(repo user.Repository, jwtSvc *auth.JWTService, log logger.Logger) *LoginUseCase {
	return &LoginUseCase{
		userRepo: repo,
		jwtSvc:   jwtSvc,
		logger:   log,
	}
}

type LoginInput struct {
	Email    string
	Password string
}

type LoginOutput struct {
	AccessToken string
}

func (uc *LoginUseCase) Execute(ctx context.Context, input LoginInput) (*LoginOutput, error) {

	u, err := uc.userRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		return nil, err
	}

	if !auth.CheckPasswordHash(input.Password, u.PasswordHash) {
		return nil, apperror.NewUnauthorized("incorrect password", nil)
	}

	token, err := uc.jwtSvc.GenerateToken(u.ID)
	if err != nil {
		uc.logger.Error("Failed to generate token", err, zap.String("user_id", u.ID.String()))
		return nil, apperror.NewInternal("failed to generate token", err)
	}

	return &LoginOutput{AccessToken: token}, nil
}
