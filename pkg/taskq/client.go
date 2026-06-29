package taskq

import (
	"context"
	"fmt"
	"go-server-starter/internal/config"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

// Client wraps an Asynq client with logging and deduplication.
type Client struct {
	*asynq.Client
	logger *zap.Logger
}

// NewClient creates a new Asynq client connected to the configured Redis.
func NewClient(cfg config.AsynQConfig, logger *zap.Logger) (*Client, error) {
	redisOpt := asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%d", cfg.RedisConfig.Host, cfg.RedisConfig.Port),
		Password: cfg.RedisConfig.Password,
		DB:       cfg.RedisConfig.DB,
	}
	return &Client{
		Client: asynq.NewClient(redisOpt),
		logger: logger,
	}, nil
}

// Enqueue enqueues a task with logging. Use EnqueueUnique for idempotent delivery.
func (c *Client) Enqueue(ctx context.Context, task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	info, err := c.Client.EnqueueContext(ctx, task, opts...)
	if err != nil {
		c.logger.Error("failed to enqueue task",
			zap.String("type", task.Type()),
			zap.Error(err),
		)
		return nil, err
	}
	c.logger.Info("task enqueued",
		zap.String("id", info.ID),
		zap.String("type", task.Type()),
		zap.String("queue", info.Queue),
	)
	return info, nil
}

// EnqueueUnique enqueues a task with deduplication.
func (c *Client) EnqueueUnique(ctx context.Context, task *asynq.Task, uniqueKey string, ttl time.Duration, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	opts = append(opts, asynq.TaskID(uniqueKey), asynq.Unique(ttl))
	return c.Enqueue(ctx, task, opts...)
}
