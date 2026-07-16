package service

import (
	"context"
	"encoding/json"
	"errors"
	"go-server-starter/internal/constant"
	cctx "go-server-starter/internal/ctx"
	"go-server-starter/internal/dto"
	"go-server-starter/internal/exception"
	"go-server-starter/internal/repo"
	"go-server-starter/pkg/redis"
	"go-server-starter/pkg/utils"

	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

type PermissionService interface {
	GetPermissionCodesByUniCode(ctx context.Context, uniCode string) ([]string, *exception.Exception)
	GetCachedPermissionCodesByUniCode(ctx context.Context, uniCode string) ([]string, *exception.Exception)
	GetMyAccess(ctx context.Context, uniCode string) (*dto.MyAccessResDto, *exception.Exception)
	GetTable(ctx context.Context, params dto.PermissionTableQueryReqDto) (*dto.PaginationResDto[[]*dto.PermissionResDto], *exception.Exception)
	DeleteAccessCache(ctx context.Context, uniCode string)
	DeleteTenantAccessCache(ctx context.Context, tenantID uint64, uniCode string)
	InvalidateTenantAccessCaches(ctx context.Context, tenantID uint64)
}

type PermissionServiceImpl struct {
	repo    repo.Repo
	redis   *redis.Client
	logger  *zap.Logger
	sfGroup singleflight.Group
}

func NewPermissionService(repo repo.Repo, redis *redis.Client, logger *zap.Logger) PermissionService {
	return &PermissionServiceImpl{repo: repo, redis: redis, logger: logger}
}

func (s *PermissionServiceImpl) getActiveUserInTenant(ctx context.Context, uniCode string) (uint64, *exception.Exception) {
	user, err := s.repo.User().GetByUniCode(ctx, uniCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, exception.Unauthorized.Append("user not found")
		}
		return 0, exception.InternalServerError.Append(err.Error())
	}
	if !user.Active {
		return 0, exception.Unauthorized.Append("user is disabled")
	}
	tenantID := cctx.GetTenantID(ctx)
	if tenantID == 0 {
		return 0, exception.Unauthorized.Append("tenant not found")
	}
	isMember := user.TenantID == tenantID
	if !isMember {
		isMember, err = s.repo.User().HasTenantMembership(ctx, user.ID, tenantID)
		if err != nil {
			return 0, exception.InternalServerError.Append(err.Error())
		}
	}
	if !isMember {
		return 0, exception.Unauthorized.Append("user is not a member of tenant")
	}
	tenant, err := s.repo.Tenant().GetByID(ctx, tenantID)
	if err != nil || tenant == nil || !tenant.Active {
		return 0, exception.Unauthorized.Append("tenant is disabled or deleted")
	}
	return user.ID, nil
}

func (s *PermissionServiceImpl) GetPermissionCodesByUniCode(ctx context.Context, uniCode string) ([]string, *exception.Exception) {
	userID, exc := s.getActiveUserInTenant(ctx, uniCode)
	if exc != nil {
		return nil, exc
	}
	codes, err := s.repo.UserRole().GetPermissionCodesByUserAndTenant(ctx, userID, cctx.GetTenantID(ctx))
	if err != nil {
		return nil, exception.InternalServerError.Append(err.Error())
	}
	return codes, nil
}

func (s *PermissionServiceImpl) GetCachedPermissionCodesByUniCode(ctx context.Context, uniCode string) ([]string, *exception.Exception) {
	cacheKey := constant.RedisKeyOfAuthPermissions(cctx.GetTenantID(ctx), uniCode)
	if data, err := s.redis.Get(ctx, cacheKey).Result(); err == nil {
		var permissions []string
		if err := json.Unmarshal([]byte(data), &permissions); err == nil {
			return permissions, nil
		}
	} else if !errors.Is(err, goredis.Nil) {
		s.logger.Warn("redis unavailable, falling back to DB", zap.Error(err))
	}

	result, err, _ := s.sfGroup.Do(cacheKey, func() (any, error) {
		permissions, exc := s.GetPermissionCodesByUniCode(ctx, uniCode)
		if exc != nil {
			return nil, exc
		}
		data, marshalErr := json.Marshal(permissions)
		if marshalErr == nil {
			if setErr := s.redis.Set(ctx, cacheKey, data, constant.REDIS_EXPIRE_OF_AUTH_PERMISSIONS).Err(); setErr != nil {
				s.logger.Warn("failed to cache permissions", zap.Error(setErr))
			}
		}
		return permissions, nil
	})
	if err != nil {
		if exc, ok := err.(*exception.Exception); ok {
			return nil, exc
		}
		return nil, exception.InternalServerError.Append(err.Error())
	}
	return result.([]string), nil
}

func (s *PermissionServiceImpl) GetMyAccess(ctx context.Context, uniCode string) (*dto.MyAccessResDto, *exception.Exception) {
	roles, exc := s.repoRoles(ctx, uniCode)
	if exc != nil {
		return nil, exc
	}
	permissions, exc := s.GetCachedPermissionCodesByUniCode(ctx, uniCode)
	if exc != nil {
		return nil, exc
	}
	return &dto.MyAccessResDto{Roles: roles, Permissions: permissions}, nil
}

func (s *PermissionServiceImpl) repoRoles(ctx context.Context, uniCode string) ([]string, *exception.Exception) {
	userID, exc := s.getActiveUserInTenant(ctx, uniCode)
	if exc != nil {
		return nil, exc
	}
	roles, err := s.repo.UserRole().GetRolesByUserAndTenant(ctx, userID, cctx.GetTenantID(ctx))
	if err != nil {
		return nil, exception.InternalServerError.Append(err.Error())
	}
	return effectiveRoleCodes(roles), nil
}

func permissionToDto(permissionID uint64, code, name, description string, enabled bool) dto.PermissionResDto {
	return dto.PermissionResDto{ID: permissionID, Code: code, Name: name, Description: description, Enabled: enabled}
}

func (s *PermissionServiceImpl) GetTable(ctx context.Context, params dto.PermissionTableQueryReqDto) (*dto.PaginationResDto[[]*dto.PermissionResDto], *exception.Exception) {
	actor := cctx.GetUserUniCodeFromContext(ctx)
	if actor == "" {
		return nil, exception.Forbidden.Append("authorization identity not found")
	}
	allowedCodes, exc := s.GetCachedPermissionCodesByUniCode(ctx, actor)
	if exc != nil {
		return nil, exc
	}
	permissions, total, err := s.repo.Permission().GetTable(ctx, params.Page, params.PageSize,
		repo.Where("code IN ?", allowedCodes), repo.Order("code ASC"),
		repo.Where("code NOT IN ?", constant.TenantManagementPermissions),
		repo.WhereAutoLike("code", params.Code), repo.WherePtrNonEmpty("enabled = ?", params.Enabled))
	if err != nil {
		return nil, exception.InternalServerError.Append(err.Error())
	}
	res := make([]*dto.PermissionResDto, len(permissions))
	for i, permission := range permissions {
		item := permissionToDto(permission.ID, permission.Code, permission.Name, permission.Description, permission.Enabled)
		res[i] = &item
	}
	return utils.AssemblePaginationResDto(res, total, params.Page, params.PageSize), nil
}

func (s *PermissionServiceImpl) DeleteAccessCache(ctx context.Context, uniCode string) {
	for _, pattern := range []string{"auth:roles:*:" + uniCode, "auth:permissions:*:" + uniCode} {
		keys, err := s.redis.Keys(ctx, pattern).Result()
		if err != nil {
			s.logger.Warn("failed to find user access cache keys", zap.String("pattern", pattern), zap.Error(err))
			continue
		}
		if len(keys) > 0 {
			if err := s.redis.Del(ctx, keys...).Err(); err != nil {
				s.logger.Warn("failed to delete user access cache", zap.String("pattern", pattern), zap.Error(err))
			}
		}
	}
}

func (s *PermissionServiceImpl) DeleteTenantAccessCache(ctx context.Context, tenantID uint64, uniCode string) {
	keys := []string{
		constant.RedisKeyOfAuthRoles(tenantID, uniCode),
		constant.RedisKeyOfAuthPermissions(tenantID, uniCode),
	}
	if err := s.redis.Del(ctx, keys...).Err(); err != nil {
		s.logger.Warn("failed to delete tenant user access cache",
			zap.Uint64("tenantID", tenantID), zap.String("uniCode", uniCode), zap.Error(err))
	}
}

func (s *PermissionServiceImpl) InvalidateTenantAccessCaches(ctx context.Context, tenantID uint64) {
	patterns := []string{
		constant.RedisKeyOfAuthRoles(tenantID, "*"),
		constant.RedisKeyOfAuthPermissions(tenantID, "*"),
	}
	for _, pattern := range patterns {
		var cursor uint64
		for {
			keys, next, err := s.redis.Scan(ctx, cursor, pattern, 100).Result()
			if err != nil {
				s.logger.Warn("failed to scan access cache keys", zap.String("pattern", pattern), zap.Error(err))
				break
			}
			if len(keys) > 0 {
				if err := s.redis.Del(ctx, keys...).Err(); err != nil {
					s.logger.Warn("failed to delete access cache", zap.String("pattern", pattern), zap.Error(err))
				}
			}
			cursor = next
			if cursor == 0 {
				break
			}
		}
	}
}
