package repo

import (
	"go-server-starter/internal/model"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type DeadLetterRepo interface {
	BaseRepo[model.DeadLetter]
}

type deadLetterRepoImpl struct {
	BaseRepo[model.DeadLetter]
}

func NewDeadLetterRepo(db *gorm.DB, logger *zap.Logger) DeadLetterRepo {
	return &deadLetterRepoImpl{
		BaseRepo: NewBaseRepo[model.DeadLetter](db, logger),
	}
}
