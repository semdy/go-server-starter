package cronjob

import (
	"context"
	"time"

	"go-server-starter/internal/repo"

	"go.uber.org/zap"
)

// Job represents a named cron job to be registered.
type Job struct {
	Name string
	Spec string
	Fn   func()
}

// Register creates a scheduler with all cron jobs registered.
// Pass the returned scheduler to app.Start and app.Shutdown.
func Register(repo repo.Repo, logger *zap.Logger) *Scheduler {
	s := New(logger)
	log := logger.Named("CRON-JOB")
	ctx := context.Background()

	jobs := []Job{
		{Name: "heartbeat", Spec: "@every 1h", Fn: heartbeat(log)},
		{Name: "purge-dead-letters", Spec: "0 0 3 * * *", Fn: purgeRetriedDeadLetters(ctx, repo, log, 30*24*time.Hour)},
	}

	for _, j := range jobs {
		s.Add(j.Name, j.Spec, j.Fn)
	}
	return s
}

// ----- job implementations -----

func heartbeat(log *zap.Logger) func() {
	return func() { log.Info("cron heartbeat") }
}

func purgeRetriedDeadLetters(ctx context.Context, repo repo.Repo, log *zap.Logger, ttl time.Duration) func() {
	return func() {
		cutoff := time.Now().Add(-ttl)
		deleted, err := repo.DeadLetter().HardDeleteRetriedBefore(ctx, cutoff)
		if err != nil {
			log.Warn("failed to purge retried dead letters", zap.Error(err))
			return
		}
		if deleted > 0 {
			log.Info("purged retried dead letters",
				zap.Int64("deleted", deleted),
				zap.Time("cutoff", cutoff),
			)
		}
	}
}
