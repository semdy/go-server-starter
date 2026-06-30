package repo

import (
	"go-server-starter/internal/model"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type TenantRepo interface {
	BaseRepo[model.Tenant]
}

type tenantRepoImpl struct {
	BaseRepo[model.Tenant]
}

func NewTenantRepo(db *gorm.DB, logger *zap.Logger) TenantRepo {
	return &tenantRepoImpl{
		BaseRepo: NewBaseRepo[model.Tenant](db, logger),
	}
}
