package redis

import (
	"context"
	"fmt"
	"go-server-starter/internal/config"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Client wraps a Redis client with a logger.
type Client struct {
	*redis.Client
	logger *zap.Logger
}

// NewClient creates a new Redis client and verifies connectivity.
func NewClient(cfg config.RedisConfig, logger *zap.Logger, ctx context.Context) (*Client, error) {
	db := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := db.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("connect to redis failed: %w", err)
	}
	return &Client{db, logger}, nil
}
