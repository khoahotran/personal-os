package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/khoahotran/personal-os/internal/domain/user"
	"github.com/khoahotran/personal-os/pkg/auth"
)

var (
	ErrInvalidCredentials = errors.New("email or password is incorrect")
)

type LoginUseCase struct {
	userRepo user.Repository
	jwtSvc   *auth.JWTService
}

func NewLoginUseCase(repo user.Repository, jwtSvc *auth.JWTService) *LoginUseCase {
	return &LoginUseCase{
		userRepo: repo,
		jwtSvc:   jwtSvc,
	}
}

type LoginInput struct {
	Email string
	Password string
}

type LoginOutput struct {
	AccessToken string
}

func (uc *LoginUseCase) Execute(ctx context.Context, input LoginInput) (*LoginOutput, error) {

	u, err := uc.userRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if !auth.CheckPasswordHash(input.Password, u.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	token, err := uc.jwtSvc.GenerateToken(u.ID)
	if err != nil {
		return nil, fmt.Errorf("cannot generate token: %w", err)
	}

	return &LoginOutput{AccessToken: token}, nil
}
