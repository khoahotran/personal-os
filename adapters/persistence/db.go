package persistence

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/khoahotran/personal-os/internal/config"
)

func NewPostgresPool(cfg config.Config) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), cfg.DB.DSN)
	if err != nil {
		return nil, fmt.Errorf("do not create connection pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database failed: %w", err)
	}

	fmt.Println("Connect PostgreSQL successfully.")
	return pool, nil
}