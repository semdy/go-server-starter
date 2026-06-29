package taskq

import (
	"context"
	"fmt"
	"go-server-starter/internal/config"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

// Client wraps an Asynq client with logging and convenience methods.
type Client struct {
	*asynq.Client
	logger *zap.Logger
}

// NewClient creates a new Asynq client connected to the configured Redis.
func NewClient(cfg config.AsynQConfig, logger *zap.Logger) (*Client, error) {
	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%d", cfg.RedisConfig.Host, cfg.RedisConfig.Port),
		Password: cfg.RedisConfig.Password,
		DB:       cfg.RedisConfig.DB,
	})
	return &Client{Client: client, logger: logger}, nil
}

// Enqueue enqueues a task with logging.
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
