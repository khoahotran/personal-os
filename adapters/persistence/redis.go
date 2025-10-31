package persistence

import (
	"context"
	"fmt"
	"github.com/khoahotran/personal-os/internal/config"
	"github.com/redis/go-redis/v9"
)

func NewRedisClient(cfg config.Config) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       0,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("can not connect Redis: %w", err)
	}

	fmt.Println("Connect Redis successfully.")
	return rdb, nil
}
