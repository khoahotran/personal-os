package persistence

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/khoahotran/personal-os/internal/domain/user"
	"github.com/khoahotran/personal-os/pkg/apperror"
	"github.com/khoahotran/personal-os/pkg/logger"
)

type postgresUserRepo struct {
	db     *pgxpool.Pool
	logger logger.Logger
}

func NewPostgresUserRepo(db *pgxpool.Pool, logger logger.Logger) user.Repository {
	return &postgresUserRepo{db: db, logger: logger}
}

func (r *postgresUserRepo) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	query := `
		SELECT id, email, name, password_hash, profile_settings
		FROM users
		WHERE email = $1
	`
	u := &user.User{}
	var profileSettingsBytes []byte

	err := r.db.QueryRow(ctx, query, email).Scan(
		&u.ID,
		&u.Email,
		&u.Name,
		&u.PasswordHash,
		&profileSettingsBytes,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.NewUnauthorized("user not found", nil)
		}
		return nil, apperror.NewInternal("error querying user", err)
	}

	if err := json.Unmarshal(profileSettingsBytes, &u.ProfileSettings); err != nil {
		r.logger.Warn("Failed to unmarshal profile_settings", zap.String("email", email), zap.Error(err))
		u.ProfileSettings = map[string]any{}
	}

	return u, nil
}
