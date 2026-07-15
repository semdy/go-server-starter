package repo

import (
	"go-server-starter/internal/model"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type PermissionRepo interface {
	BaseRepo[model.Permission]
	WithTx(tx *gorm.DB) PermissionRepo
}

type PermissionRepoImpl struct {
	BaseRepo[model.Permission]
	db     *gorm.DB
	logger *zap.Logger
}

func NewPermissionRepo(db *gorm.DB, logger *zap.Logger) PermissionRepo {
	return &PermissionRepoImpl{
		BaseRepo: NewBaseRepo[model.Permission](db, logger),
		db:       db,
		logger:   logger,
	}
}

func (r *PermissionRepoImpl) WithTx(tx *gorm.DB) PermissionRepo {
	return &PermissionRepoImpl{
		BaseRepo: NewBaseRepo[model.Permission](tx, r.logger),
		db:       tx,
		logger:   r.logger,
	}
}
