package persistence

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/khoahotran/personal-os/internal/domain/user"
)

type postgresUserRepo struct {
	db *pgxpool.Pool
}

func NewPostgresUserRepo(db *pgxpool.Pool) user.Repository {
	return &postgresUserRepo{db: db}
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
			return nil, errors.New("user not exist")
		}
		return nil, fmt.Errorf("error when query user: %w", err)
	}

	return u, nil
}
