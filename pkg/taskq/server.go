package taskq

import (
	"context"
	"fmt"
	"sync/atomic"

	"go-server-starter/internal/config"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

// ServerConfig holds server-level configuration for the task worker.
type ServerConfig struct {
	Concurrency int            // max concurrent task processors
	Queues      map[string]int // queue name → priority weight
}

// Server wraps an Asynq server with pre-configured mux, error handling, and alerting.
type alerterPtr struct {
	v atomic.Pointer[Alerter]
}

func (a *alerterPtr) get() Alerter {
	if p := a.v.Load(); p != nil {
		return *p
	}
	return noopAlerter{}
}

func (a *alerterPtr) set(impl Alerter) {
	a.v.Store(&impl)
}

type Server struct {
	*asynq.Server
	mux         *asynq.ServeMux
	logger      *zap.Logger
	concurrency int
	alerter     *alerterPtr
}

// NewServer creates a new Asynq server connected to the configured Redis.
// Pass nil for alerter to use the default (no-op) implementation.
func NewServer(redisConfig config.AsynQConfig, cfg ServerConfig, logger *zap.Logger, alerter Alerter) *Server {
	logger = logger.Named("TASKQ")
	a := &alerterPtr{}
	if alerter != nil {
		a.set(alerter)
	}
	srv := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     fmt.Sprintf("%s:%d", redisConfig.RedisConfig.Host, redisConfig.RedisConfig.Port),
			Password: redisConfig.RedisConfig.Password,
			DB:       redisConfig.RedisConfig.DB,
		},
		asynq.Config{
			Concurrency: cfg.Concurrency,
			Queues:      cfg.Queues,
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				retried, _ := asynq.GetRetryCount(ctx)
				maxRetry, _ := asynq.GetMaxRetry(ctx)
				if retried >= maxRetry {
					// All retries exhausted — escalate via alerter
					logger.Error("task exhausted all retries",
						zap.String("type", task.Type()),
						zap.String("id", task.ResultWriter().TaskID()),
						zap.Int("attempt", retried),
						zap.Int("maxRetry", maxRetry),
						zap.Error(err),
					)
					a.get().Alert(ctx, AlertInfo{
						TaskType: task.Type(),
						TaskID:   task.ResultWriter().TaskID(),
						TenantID: TenantIDFromPayload(task.Payload()),
						Payload:  task.Payload(),
						Error:    err.Error(),
						Attempt:  retried,
						MaxRetry: maxRetry,
					})
				} else {
					logger.Warn("task processing failed, will retry",
						zap.String("type", task.Type()),
						zap.String("id", task.ResultWriter().TaskID()),
						zap.Int("attempt", retried),
						zap.Int("maxRetry", maxRetry),
						zap.Error(err),
					)
				}
			}),
		},
	)
	return &Server{Server: srv, mux: asynq.NewServeMux(), logger: logger, concurrency: cfg.Concurrency, alerter: a}
}

// SetAlerter replaces the active alerter. Safe for concurrent use.
func (s *Server) SetAlerter(impl Alerter) {
	s.alerter.set(impl)
}

// Handle registers a task handler for the given pattern.
func (s *Server) Handle(pattern string, handler asynq.Handler) {
	s.mux.Handle(pattern, handler)
}

// HandleFunc registers a task handler function for the given pattern.
func (s *Server) HandleFunc(pattern string, handler func(context.Context, *asynq.Task) error) {
	s.mux.HandleFunc(pattern, handler)
}

// Start starts the task worker. Blocks until the server is stopped.
func (s *Server) Start() error {
	s.logger.Info("asynq worker server starting",
		zap.Int("concurrency", s.concurrency),
	)
	return s.Server.Start(s.mux)
}

// Shutdown gracefully stops the task worker, waiting for in-flight tasks to complete.
func (s *Server) Shutdown() {
	s.Server.Shutdown()
	s.logger.Info("asynq worker server stopped")
}
