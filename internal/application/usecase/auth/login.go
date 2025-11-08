package auth

import (
	"context"
	"errors"

	"github.com/khoahotran/personal-os/internal/domain/user"
	"github.com/khoahotran/personal-os/pkg/apperror"
	"github.com/khoahotran/personal-os/pkg/auth"
	"github.com/khoahotran/personal-os/pkg/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
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

var tracer = otel.Tracer("auth_usecase")

func (uc *LoginUseCase) Execute(ctx context.Context, input LoginInput) (*LoginOutput, error) {

	ctx, span := tracer.Start(ctx, "Execute")
	defer span.End()

	u, err := uc.userRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	if !auth.CheckPasswordHash(input.Password, u.PasswordHash) {
		err := apperror.NewUnauthorized("incorrect password", nil)
		span.RecordError(err)
		return nil, err
	}

	token, err := uc.jwtSvc.GenerateToken(u.ID)
	if err != nil {
		uc.logger.Error("Failed to generate token", err, zap.String("user_id", u.ID.String()))
		err = apperror.NewInternal("failed to generate token", err)
		span.RecordError(err)
		return nil, err
	}
	span.SetAttributes(attribute.String("user_id", u.ID.String()))
	return &LoginOutput{AccessToken: token}, nil
}
