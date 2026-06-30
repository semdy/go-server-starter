package cronjob

import (
	"context"
	"sync"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// Scheduler wraps robfig/cron with named jobs and graceful shutdown.
type Scheduler struct {
	cron   *cron.Cron
	logger *zap.Logger
	mu     sync.Mutex
	ids    map[string]cron.EntryID
}

// New creates a new cron scheduler.
func New(logger *zap.Logger) *Scheduler {
	return &Scheduler{
		cron: cron.New(
			cron.WithSeconds(),                       // 6-field: sec min hour dom month dow
			cron.WithLogger(cronLogger{logger.Named("CRON")}),
		),
		logger: logger.Named("CRON"),
		ids:    make(map[string]cron.EntryID),
	}
}

// Add adds a named job with a cron expression.
// Supports 6-field (sec min hour dom month dow) and @every 1h30m.
func (s *Scheduler) Add(name, spec string, cmd func()) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id, err := s.cron.AddFunc(spec, cmd)
	if err != nil {
		s.logger.Error("failed to add cron job", zap.String("name", name), zap.Error(err))
		return
	}
	s.ids[name] = id
	s.logger.Info("cron job registered", zap.String("name", name), zap.String("spec", spec), zap.Int("id", int(id)))
}

// Start begins executing jobs. Non-blocking.
func (s *Scheduler) Start() {
	s.logger.Info("cron scheduler starting", zap.Int("jobs", len(s.ids)))
	s.cron.Start()
}

// Stop waits for running jobs to complete, then returns.
func (s *Scheduler) Stop() context.Context {
	ctx := s.cron.Stop()
	s.logger.Info("cron scheduler stopped")
	return ctx
}

// Remove removes a named job.
func (s *Scheduler) Remove(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if id, ok := s.ids[name]; ok {
		s.cron.Remove(id)
		delete(s.ids, name)
		s.logger.Info("cron job removed", zap.String("name", name))
	}
}

// cronLogger adapts zap to robfig/cron's Logger interface.
type cronLogger struct {
	*zap.Logger
}

func (l cronLogger) Info(msg string, keysAndValues ...interface{}) {
	l.Logger.Sugar().Infow(msg, keysAndValues...)
}
func (l cronLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	l.Logger.Sugar().Errorw(msg, append([]interface{}{"error", err}, keysAndValues...)...)
}
