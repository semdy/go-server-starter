package repo

import (
	"context"
	"go-server-starter/internal/model"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type DeadLetterRepo interface {
	BaseRepo[model.DeadLetter]
	// HardDeleteRetriedBefore permanently deletes retried dead letters older than the cutoff.
	HardDeleteRetriedBefore(ctx context.Context, cutoff time.Time) (int64, error)
}

type deadLetterRepoImpl struct {
	BaseRepo[model.DeadLetter]
	db *gorm.DB
}

func NewDeadLetterRepo(db *gorm.DB, logger *zap.Logger) DeadLetterRepo {
	return &deadLetterRepoImpl{
		BaseRepo: NewBaseRepo[model.DeadLetter](db, logger),
		db:       db,
	}
}

func (r *deadLetterRepoImpl) HardDeleteRetriedBefore(ctx context.Context, cutoff time.Time) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("is_retried = ? AND retried_at < ?", true, cutoff).
		Unscoped().
		Delete(&model.DeadLetter{})
	return result.RowsAffected, result.Error
}
