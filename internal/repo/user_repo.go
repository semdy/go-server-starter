package repo

import (
	"context"
	"errors"
	"fmt"
	"go-server-starter/internal/model"
	"go-server-starter/pkg/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UserRepo interface {
	BaseRepo[model.User]
	WithTx(tx *gorm.DB) UserRepo
	GenerateUniCode(ctx context.Context) (string, error)
	GetIDByUniCode(ctx context.Context, uniCode string) (uint64, error)
	GetByUniCode(ctx context.Context, uniCode string) (*model.User, error)
	AddTenantMembership(ctx context.Context, userID uint64, tenantID uint64) error
	HasTenantMembership(ctx context.Context, userID uint64, tenantID uint64) (bool, error)
	GetTenantsByUserID(ctx context.Context, userID uint64) ([]*model.Tenant, error)
}

type UserRepoImpl struct {
	BaseRepo[model.User]
	db     *gorm.DB
	logger *zap.Logger
}

func NewUserRepo(db *gorm.DB, logger *zap.Logger) UserRepo {
	return &UserRepoImpl{
		BaseRepo: NewBaseRepo[model.User](db, logger),
		db:       db,
		logger:   logger,
	}
}

func (r *UserRepoImpl) WithTx(tx *gorm.DB) UserRepo {
	return &UserRepoImpl{
		BaseRepo: NewBaseRepo[model.User](tx, r.logger),
		db:       tx,
		logger:   r.logger,
	}
}

func (r *UserRepoImpl) GenerateUniCode(ctx context.Context) (string, error) {
	for {
		uniCode := utils.RandomUserCode()
		exists, err := r.GetOne(ctx, Where("uni_code = ?", uniCode))
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return "", fmt.Errorf("query user by uni code failed: %w", err)
		}
		if exists == nil {
			return uniCode, nil
		}
	}
}

func (r *UserRepoImpl) GetIDByUniCode(ctx context.Context, uniCode string) (uint64, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Where("uni_code = ?", uniCode).First(&user).Error; err != nil {
		return 0, err
	}
	return user.ID, nil
}

func (r *UserRepoImpl) GetByUniCode(ctx context.Context, uniCode string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Where("uni_code = ?", uniCode).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepoImpl) AddTenantMembership(ctx context.Context, userID uint64, tenantID uint64) error {
	return r.db.WithContext(ctx).
		Exec("INSERT IGNORE INTO user_tenant_refs (user_id, tenant_id) VALUES (?, ?)", userID, tenantID).
		Error
}

func (r *UserRepoImpl) HasTenantMembership(ctx context.Context, userID uint64, tenantID uint64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("user_tenant_refs").
		Where("user_id = ? AND tenant_id = ?", userID, tenantID).
		Count(&count).Error
	return count > 0, err
}

func (r *UserRepoImpl) GetTenantsByUserID(ctx context.Context, userID uint64) ([]*model.Tenant, error) {
	tenants := make([]*model.Tenant, 0)
	err := r.db.WithContext(ctx).
		Table("tenants AS t").
		Select("t.*").
		Joins("JOIN user_tenant_refs AS utr ON utr.tenant_id = t.id").
		Where("utr.user_id = ? AND t.active = ? AND t.deleted_at IS NULL", userID, true).
		Order("t.name ASC, t.id ASC").
		Find(&tenants).Error
	return tenants, err
}
