package repo

import (
	"context"
	"go-server-starter/internal/model"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UserRoleRepo interface {
	BaseRepo[model.UserRole]
	WithTx(tx *gorm.DB) UserRoleRepo
	GetRolesByUserAndTenant(ctx context.Context, userID, tenantID uint64) ([]*model.UserRole, error)
	GetPermissionCodesByUserAndTenant(ctx context.Context, userID, tenantID uint64) ([]string, error)
	ReplaceUserRoles(ctx context.Context, userID, tenantID uint64, roleIDs []uint64) error
	AddUserRole(ctx context.Context, userID, tenantID, roleID uint64) error
	ReplaceRolePermissions(ctx context.Context, roleID uint64, permissionIDs []uint64) error
	SetRolePermission(ctx context.Context, roleID, permissionID uint64, checked bool) error
}

type UserRoleRepoImpl struct {
	BaseRepo[model.UserRole]
	db     *gorm.DB
	logger *zap.Logger
}

func NewUserRoleRepo(db *gorm.DB, logger *zap.Logger) UserRoleRepo {
	return &UserRoleRepoImpl{
		BaseRepo: NewBaseRepo[model.UserRole](db, logger),
		db:       db,
		logger:   logger,
	}
}

func (r *UserRoleRepoImpl) WithTx(tx *gorm.DB) UserRoleRepo {
	return &UserRoleRepoImpl{
		BaseRepo: NewBaseRepo[model.UserRole](tx, r.logger),
		db:       tx,
		logger:   r.logger,
	}
}

func (r *UserRoleRepoImpl) GetRolesByUserAndTenant(ctx context.Context, userID, tenantID uint64) ([]*model.UserRole, error) {
	roles := make([]*model.UserRole, 0)
	err := r.db.WithContext(ctx).
		Table("user_roles AS r").
		Select("r.*").
		Joins("JOIN user_tenant_role_refs AS utr ON utr.role_id = r.id").
		Where("utr.user_id = ? AND utr.tenant_id = ? AND (r.tenant_id = 0 OR r.tenant_id = utr.tenant_id) AND r.enabled = ? AND r.deleted_at IS NULL", userID, tenantID, true).
		Order("r.built_in DESC, r.code ASC").
		Find(&roles).Error
	return roles, err
}

func (r *UserRoleRepoImpl) GetPermissionCodesByUserAndTenant(ctx context.Context, userID, tenantID uint64) ([]string, error) {
	var codes []string
	err := r.db.WithContext(ctx).
		Table("permissions AS p").
		Distinct("p.code").
		Joins("JOIN role_permission_refs AS rp ON rp.permission_id = p.id").
		Joins("JOIN user_roles AS r ON r.id = rp.role_id AND r.deleted_at IS NULL AND r.enabled = ?", true).
		Joins("JOIN user_tenant_role_refs AS utr ON utr.role_id = r.id").
		Where("utr.user_id = ? AND utr.tenant_id = ? AND (r.tenant_id = 0 OR r.tenant_id = utr.tenant_id) AND p.enabled = ? AND p.deleted_at IS NULL", userID, tenantID, true).
		Order("p.code ASC").
		Pluck("p.code", &codes).Error
	return codes, err
}

func (r *UserRoleRepoImpl) ReplaceUserRoles(ctx context.Context, userID, tenantID uint64, roleIDs []uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM user_tenant_role_refs WHERE user_id = ? AND tenant_id = ?", userID, tenantID).Error; err != nil {
			return err
		}
		for _, roleID := range roleIDs {
			if err := tx.Exec("INSERT INTO user_tenant_role_refs (user_id, tenant_id, role_id) VALUES (?, ?, ?)", userID, tenantID, roleID).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *UserRoleRepoImpl) AddUserRole(ctx context.Context, userID, tenantID, roleID uint64) error {
	return r.db.WithContext(ctx).
		Exec("INSERT IGNORE INTO user_tenant_role_refs (user_id, tenant_id, role_id) VALUES (?, ?, ?)", userID, tenantID, roleID).
		Error
}

func (r *UserRoleRepoImpl) ReplaceRolePermissions(ctx context.Context, roleID uint64, permissionIDs []uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM role_permission_refs WHERE role_id = ?", roleID).Error; err != nil {
			return err
		}
		for _, permissionID := range permissionIDs {
			if err := tx.Exec("INSERT INTO role_permission_refs (role_id, permission_id) VALUES (?, ?)", roleID, permissionID).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *UserRoleRepoImpl) SetRolePermission(ctx context.Context, roleID, permissionID uint64, checked bool) error {
	if checked {
		return r.db.WithContext(ctx).
			Exec("INSERT IGNORE INTO role_permission_refs (role_id, permission_id) VALUES (?, ?)", roleID, permissionID).
			Error
	}
	return r.db.WithContext(ctx).
		Exec("DELETE FROM role_permission_refs WHERE role_id = ? AND permission_id = ?", roleID, permissionID).
		Error
}
