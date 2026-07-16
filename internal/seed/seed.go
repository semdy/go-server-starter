package seed

import (
	"context"
	"errors"
	"fmt"

	"go-server-starter/internal/constant"
	"go-server-starter/internal/enum"
	"go-server-starter/internal/model"
	"go-server-starter/internal/repo"
	"go-server-starter/pkg/redis"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Seed interface {
	Run() error
}

type seed struct {
	repo   repo.Repo
	redis  *redis.Client
	logger *zap.Logger
}

func NewSeed(repo repo.Repo, redis *redis.Client, logger *zap.Logger) Seed {
	return &seed{repo: repo, redis: redis, logger: logger}
}

func (s *seed) Run() error {
	if err := s.SeedUserRole(); err != nil {
		return err
	}
	if err := s.SeedDefaultTenant(); err != nil {
		return err
	}
	if err := s.SeedPermissions(); err != nil {
		return err
	}
	return s.clearAccessCaches()
}

func (s *seed) clearAccessCaches() error {
	ctx := context.Background()
	var deleted int64
	for _, pattern := range []string{"auth:roles:*", "auth:permissions:*"} {
		var cursor uint64
		for {
			keys, next, err := s.redis.Scan(ctx, cursor, pattern, 100).Result()
			if err != nil {
				return fmt.Errorf("scan access cache keys with pattern %q: %w", pattern, err)
			}
			if len(keys) > 0 {
				count, err := s.redis.Del(ctx, keys...).Result()
				if err != nil {
					return fmt.Errorf("delete access cache keys with pattern %q: %w", pattern, err)
				}
				deleted += count
			}
			cursor = next
			if cursor == 0 {
				break
			}
		}
	}
	s.logger.Info("cleared access caches after seed", zap.Int64("count", deleted))
	return nil
}

func (s *seed) SeedDefaultTenant() error {
	ctx := context.Background()
	existing, err := s.repo.Tenant().GetOne(ctx, repo.Where("code = ?", "default"))
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if existing != nil {
		return nil
	}
	t := &model.Tenant{Name: "Default", Code: "default", Active: true}
	if err := s.repo.Tenant().Create(ctx, t); err != nil {
		return err
	}
	s.logger.Info("seeded default tenant")
	return nil
}

func (s *seed) SeedUserRole() error {
	var ctx = context.Background()
	var roles = []enum.RoleCode{
		enum.RoleCodeSuperAdmin,
		enum.RoleCodeAdmin,
		enum.RoleCodeGuest,
		enum.RoleCodeUser,
		enum.RoleCodeUserVip,
		enum.RoleCodeUserSvip,
	}

	// 查询数据库中已存在的角色
	existingRoles, err := s.repo.UserRole().GetMany(ctx)
	if err != nil {
		s.logger.Error("failed to query existing roles", zap.Error(err))
		return err
	}

	// 创建已存在角色代码的 map，用于快速查找
	existingRoleMap := make(map[string]bool)
	for _, role := range existingRoles {
		existingRoleMap[role.Code] = true
	}

	// 过滤出需要插入的角色
	var newRoles []*model.UserRole

	for _, roleCode := range roles {
		if !existingRoleMap[roleCode.String()] {
			newRoles = append(newRoles, &model.UserRole{
				TenantID: 0,
				Code:     roleCode.String(),
				Name:     roleCode.String(),
				BuiltIn:  true,
				Enabled:  true,
			})
		}
	}

	// 如果有需要插入的角色，批量插入
	if len(newRoles) > 0 {
		if err := s.repo.UserRole().CreateBatch(ctx, newRoles); err != nil {
			s.logger.Error("failed to insert roles", zap.Error(err))
			return err
		}
		s.logger.Info("successfully seeded roles", zap.Int("count", len(newRoles)))
	} else {
		s.logger.Info("all roles already exist, no need to seed")
	}

	return nil
}

func (s *seed) SeedPermissions() error {
	ctx := context.Background()
	for _, definition := range constant.BuiltInPermissions {
		permission, err := s.repo.Permission().GetOne(ctx, repo.Where("code = ?", definition.Code))
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if permission == nil {
			permission = &model.Permission{Code: definition.Code, Name: definition.Name, Enabled: true}
			if err := s.repo.Permission().Create(ctx, permission); err != nil {
				return err
			}
		}
	}

	allPermissions, err := s.repo.Permission().GetMany(ctx, repo.Where("enabled = ?", true))
	if err != nil {
		return err
	}
	permissionIDs := make(map[string]uint64, len(allPermissions))
	allIDs := make([]uint64, 0, len(allPermissions))
	for _, permission := range allPermissions {
		permissionIDs[permission.Code] = permission.ID
		allIDs = append(allIDs, permission.ID)
	}

	superAdmin, err := s.repo.UserRole().GetOne(ctx, repo.Where("tenant_id = 0 AND code = ?", enum.RoleCodeSuperAdmin.String()))
	if err != nil {
		return err
	}
	if err := s.repo.UserRole().ReplaceRolePermissions(ctx, superAdmin.ID, allIDs); err != nil {
		return err
	}

	admin, err := s.repo.UserRole().GetOne(ctx, repo.Where("tenant_id = 0 AND code = ?", enum.RoleCodeAdmin.String()))
	if err != nil {
		return err
	}
	adminIDs := make([]uint64, 0, len(constant.AdminPermissions))
	for _, code := range constant.AdminPermissions {
		if id, ok := permissionIDs[code]; ok {
			adminIDs = append(adminIDs, id)
		}
	}
	if err := s.repo.UserRole().ReplaceRolePermissions(ctx, admin.ID, adminIDs); err != nil {
		return err
	}

	s.logger.Info("seeded permissions and built-in role mappings")
	return nil
}
