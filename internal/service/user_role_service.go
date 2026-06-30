package service

import (
	"context"
	"encoding/json"
	"errors"
	"go-server-starter/internal/constant"
	"go-server-starter/internal/dto"
	"go-server-starter/internal/enum"
	"go-server-starter/internal/exception"
	"go-server-starter/internal/model"
	"go-server-starter/internal/repo"
	"go-server-starter/pkg/redis"
	"go-server-starter/pkg/utils"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

type UserRoleService interface {
	GetRolesCodeByUniCode(ctx context.Context, uniCode string) ([]enum.RoleCode, *exception.Exception)
	GetCachedRolesCodeByUniCode(ctx context.Context, uniCode string) ([]enum.RoleCode, *exception.Exception)
	GetByID(ctx context.Context, id uint64) (*dto.UserRoleResDto, *exception.Exception)
	GetTable(ctx context.Context, params dto.UserRoleTableQueryReqDto) (*dto.PaginationResDto[[]*dto.UserRoleResDto], *exception.Exception)
	Create(ctx context.Context, params dto.UserRoleCreateReqDto) (*dto.UserRoleResDto, *exception.Exception)
	Update(ctx context.Context, id uint64, params dto.UserRoleUpdateReqDto) (*dto.UserRoleResDto, *exception.Exception)
	Delete(ctx context.Context, id uint64) *exception.Exception
}

type UserRoleServiceImpl struct {
	repo    repo.Repo
	redis   *redis.Client
	logger  *zap.Logger
	sfGroup singleflight.Group // deduplicates concurrent cache-miss DB queries
}

func NewUserRoleService(repo repo.Repo, redis *redis.Client, logger *zap.Logger) UserRoleService {
	return &UserRoleServiceImpl{
		repo:   repo,
		redis:  redis,
		logger: logger,
	}
}

func (s *UserRoleServiceImpl) GetRolesCodeByUniCode(ctx context.Context, uniCode string) ([]enum.RoleCode, *exception.Exception) {
	roles, err := s.repo.User().GetRolesByUniCode(ctx, uniCode)
	if err != nil {
		s.logger.Error("get roles code by uni code failed", zap.String("uniCode", uniCode), zap.Error(err))
		return nil, exception.InternalServerError.Append(err.Error())
	}
	rolesCode := make([]enum.RoleCode, len(roles))
	for i, role := range roles {
		rolesCode[i] = role.Code
	}
	return rolesCode, nil
}

func (s *UserRoleServiceImpl) GetCachedRolesCodeByUniCode(ctx context.Context, uniCode string) ([]enum.RoleCode, *exception.Exception) {
	cacheKey := constant.RedisKeyOfAuthRoles(uniCode)

	dataStr, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		// Cache hit — deserialize and return
		var roles []enum.RoleCode
		if err := json.Unmarshal([]byte(dataStr), &roles); err != nil {
			s.logger.Error("unmarshal roles code by uni code failed", zap.String("uniCode", uniCode), zap.Error(err))
			return nil, exception.InternalServerError.Append(err.Error())
		}
		return roles, nil
	}

	if !errors.Is(err, goredis.Nil) {
		s.logger.Warn("redis unavailable, falling back to DB",
			zap.String("uniCode", uniCode), zap.Error(err))
		// Fall through to singleflight — don't bypass it
	}

	// Cache miss or Redis unavailable — use singleflight to deduplicate concurrent DB queries
	result, sfErr, _ := s.sfGroup.Do(cacheKey, func() (any, error) {
		roles, exc := s.GetRolesCodeByUniCode(ctx, uniCode)
		if exc != nil {
			return nil, exc // *exception.Exception implements error
		}

		rolesJSON, err := json.Marshal(roles)
		if err != nil {
			s.logger.Error("failed to marshal roles for cache", zap.String("uniCode", uniCode), zap.Error(err))
			return nil, err
		}

		if setErr := s.redis.Set(ctx, cacheKey, rolesJSON, constant.REDIS_EXPIRE_OF_AUTH_ROLES).Err(); setErr != nil {
			s.logger.Warn("failed to set role cache", zap.String("uniCode", uniCode), zap.Error(setErr))
			return nil, err
		}

		return roles, nil
	})

	if sfErr != nil {
		// singleflight returned our *exception.Exception as an error
		if exc, ok := sfErr.(*exception.Exception); ok {
			return nil, exc
		}
		return nil, exception.InternalServerError.Append(sfErr.Error())
	}

	return result.([]enum.RoleCode), nil
}

// invalidateAllRoleCaches removes all role cache entries from Redis.
// Called after any role mutation (create/update/delete) to prevent stale cache reads.
func (s *UserRoleServiceImpl) invalidateAllRoleCaches(ctx context.Context) {
	pattern := "auth:roles:*"
	var cursor uint64
	var totalDeleted int64

	for {
		keys, nextCursor, err := s.redis.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			s.logger.Warn("failed to scan role cache keys for invalidation", zap.Error(err))
			return
		}
		if len(keys) > 0 {
			deleted, err := s.redis.Del(ctx, keys...).Result()
			if err != nil {
				s.logger.Warn("failed to delete role cache keys", zap.Error(err))
			}
			totalDeleted += deleted
		}
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	if totalDeleted > 0 {
		s.logger.Info("invalidated role caches", zap.Int64("count", totalDeleted))
	}
}

// toUserRoleResDto converts a model.UserRole to a response DTO.
func toUserRoleResDto(role *model.UserRole) *dto.UserRoleResDto {
	return &dto.UserRoleResDto{
		ID:        role.ID,
		Code:      role.Code.String(),
		Enabled:   role.Enabled,
		CreatedAt: role.CreatedAt.Format(time.RFC3339),
		UpdatedAt: role.UpdatedAt.Format(time.RFC3339),
	}
}

func (s *UserRoleServiceImpl) GetByID(ctx context.Context, id uint64) (*dto.UserRoleResDto, *exception.Exception) {
	role, err := s.repo.UserRole().GetByID(ctx, id)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, exception.InternalServerError.Append(err.Error())
	}
	if role == nil {
		return nil, exception.UserRoleNotFound
	}
	return toUserRoleResDto(role), nil
}

func (s *UserRoleServiceImpl) GetTable(ctx context.Context, params dto.UserRoleTableQueryReqDto) (*dto.PaginationResDto[[]*dto.UserRoleResDto], *exception.Exception) {
	opts := []repo.QueryOption{
		repo.Order("id ASC"),
		repo.WherePtrNonEmpty("code = ?", params.Code),
		repo.WherePtrNonEmpty("enabled = ?", params.Enabled),
	}
	roles, total, err := s.repo.UserRole().GetTable(ctx, params.Page, params.PageSize, opts...)
	if err != nil {
		return nil, exception.InternalServerError.Append(err.Error())
	}
	res := make([]*dto.UserRoleResDto, len(roles))
	for i, role := range roles {
		res[i] = toUserRoleResDto(role)
	}
	return utils.AssemblePaginationResDto(res, total, params.Page, params.PageSize), nil
}

func (s *UserRoleServiceImpl) Create(ctx context.Context, params dto.UserRoleCreateReqDto) (*dto.UserRoleResDto, *exception.Exception) {
	code, err := enum.ParseRoleCode(params.Code)
	if err != nil {
		return nil, exception.InvalidParam.Append("invalid role code: " + params.Code)
	}

	// Check if role already exists
	existing, err := s.repo.UserRole().GetOne(ctx, repo.Where("code = ?", code))
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, exception.InternalServerError.Append(err.Error())
	}
	if existing != nil {
		return nil, exception.UserRoleAlreadyExists
	}

	enabled := true
	if params.Enabled != nil {
		enabled = *params.Enabled
	}

	role := &model.UserRole{Code: code, Enabled: enabled}
	if err := s.repo.UserRole().Create(ctx, role); err != nil {
		return nil, exception.InternalServerError.Append(err.Error())
	}

	// Invalidate role caches so any cached role lists reflect the new role
	s.invalidateAllRoleCaches(ctx)

	return toUserRoleResDto(role), nil
}

func (s *UserRoleServiceImpl) Update(ctx context.Context, id uint64, params dto.UserRoleUpdateReqDto) (*dto.UserRoleResDto, *exception.Exception) {
	role, err := s.repo.UserRole().GetByID(ctx, id)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, exception.InternalServerError.Append(err.Error())
	}
	if role == nil {
		return nil, exception.UserRoleNotFound
	}

	if params.Code != nil {
		code, err := enum.ParseRoleCode(*params.Code)
		if err != nil {
			return nil, exception.InvalidParam.Append("invalid role code: " + *params.Code)
		}
		role.Code = code
	}
	if params.Enabled != nil {
		role.Enabled = *params.Enabled
	}

	if err := s.repo.UserRole().UpdateByZeroFields(ctx, id, role); err != nil {
		return nil, exception.InternalServerError.Append(err.Error())
	}

	// Invalidate caches — role code or enabled status may have changed
	s.invalidateAllRoleCaches(ctx)

	return toUserRoleResDto(role), nil
}

func (s *UserRoleServiceImpl) Delete(ctx context.Context, id uint64) *exception.Exception {
	if err := s.repo.UserRole().SoftDelete(ctx, id); err != nil {
		return nil
	}

	// Invalidate caches — deleted role should no longer appear in user role lists
	s.invalidateAllRoleCaches(ctx)

	return nil
}
