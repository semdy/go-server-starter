package service

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"time"

	"go-server-starter/internal/constant"
	cctx "go-server-starter/internal/ctx"
	"go-server-starter/internal/dto"
	"go-server-starter/internal/enum"
	"go-server-starter/internal/exception"
	"go-server-starter/internal/model"
	"go-server-starter/internal/repo"
	"go-server-starter/pkg/redis"
	"go-server-starter/pkg/utils"

	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

var roleCodePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{1,49}$`)

type UserRoleService interface {
	GetRolesCodeByUniCode(ctx context.Context, uniCode string) ([]string, *exception.Exception)
	GetCachedRolesCodeByUniCode(ctx context.Context, uniCode string) ([]string, *exception.Exception)
	GetByID(ctx context.Context, id uint64) (*dto.UserRoleResDto, *exception.Exception)
	GetTable(ctx context.Context, params dto.UserRoleTableQueryReqDto) (*dto.PaginationResDto[[]*dto.UserRoleResDto], *exception.Exception)
	Create(ctx context.Context, params dto.UserRoleCreateReqDto) (*dto.UserRoleResDto, *exception.Exception)
	Update(ctx context.Context, id uint64, params dto.UserRoleUpdateReqDto) (*dto.UserRoleResDto, *exception.Exception)
	SetPermissions(ctx context.Context, id uint64, params dto.UserRoleSetPermissionsReqDto) (*dto.UserRoleResDto, *exception.Exception)
	Delete(ctx context.Context, id uint64) *exception.Exception
	InvalidateTenantAccessCaches(ctx context.Context, tenantID uint64)
	DeleteAccessCache(ctx context.Context, uniCode string)
	DeleteTenantAccessCache(ctx context.Context, tenantID uint64, uniCode string)
}

type UserRoleServiceImpl struct {
	repo    repo.Repo
	redis   *redis.Client
	access  PermissionService
	logger  *zap.Logger
	sfGroup singleflight.Group
}

func NewUserRoleService(repo repo.Repo, redis *redis.Client, access PermissionService, logger *zap.Logger) UserRoleService {
	return &UserRoleServiceImpl{repo: repo, redis: redis, access: access, logger: logger}
}

func (s *UserRoleServiceImpl) activeUserID(ctx context.Context, uniCode string) (uint64, *exception.Exception) {
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
	isMember := user.TenantID == tenantID
	if !isMember {
		isMember, err = s.repo.User().HasTenantMembership(ctx, user.ID, tenantID)
	}
	if err != nil {
		return 0, exception.InternalServerError.Append(err.Error())
	}
	if tenantID == 0 || !isMember {
		return 0, exception.Unauthorized.Append("user is not a member of tenant")
	}
	return user.ID, nil
}

func (s *UserRoleServiceImpl) GetRolesCodeByUniCode(ctx context.Context, uniCode string) ([]string, *exception.Exception) {
	userID, exc := s.activeUserID(ctx, uniCode)
	if exc != nil {
		return nil, exc
	}
	roles, err := s.repo.UserRole().GetRolesByUserAndTenant(ctx, userID, cctx.GetTenantID(ctx))
	if err != nil {
		return nil, exception.InternalServerError.Append(err.Error())
	}
	return effectiveRoleCodes(roles), nil
}

func effectiveRoleCodes(roles []*model.UserRole) []string {
	codes := make([]string, 0, len(roles))
	for _, role := range roles {
		// A legacy custom role using a reserved built-in code must never satisfy
		// role-based platform checks such as super_admin.
		if !role.BuiltIn && enum.RoleCode(role.Code).IsValid() {
			continue
		}
		codes = append(codes, role.Code)
	}
	return codes
}

func (s *UserRoleServiceImpl) GetCachedRolesCodeByUniCode(ctx context.Context, uniCode string) ([]string, *exception.Exception) {
	// A nil Redis client is useful in isolated unit tests and should not disable
	// authorization. Production always injects the configured client.
	if s.redis == nil {
		return s.GetRolesCodeByUniCode(ctx, uniCode)
	}

	cacheKey := constant.RedisKeyOfAuthRoles(cctx.GetTenantID(ctx), uniCode)
	if data, err := s.redis.Get(ctx, cacheKey).Result(); err == nil {
		var roles []string
		if err := json.Unmarshal([]byte(data), &roles); err == nil {
			return roles, nil
		}
		// Treat malformed cache data as a miss. The DB result below overwrites it.
		if s.logger != nil {
			s.logger.Warn("invalid role cache, falling back to DB", zap.String("key", cacheKey))
		}
	} else if !errors.Is(err, goredis.Nil) && s.logger != nil {
		s.logger.Warn("redis unavailable, falling back to DB", zap.Error(err))
	}

	result, err, _ := s.sfGroup.Do(cacheKey, func() (any, error) {
		roles, exc := s.GetRolesCodeByUniCode(ctx, uniCode)
		if exc != nil {
			return nil, exc
		}
		data, marshalErr := json.Marshal(roles)
		if marshalErr == nil {
			if setErr := s.redis.Set(ctx, cacheKey, data, constant.REDIS_EXPIRE_OF_AUTH_ROLES).Err(); setErr != nil && s.logger != nil {
				s.logger.Warn("failed to cache roles", zap.Error(setErr))
			}
		}
		return roles, nil
	})
	if err != nil {
		if exc, ok := err.(*exception.Exception); ok {
			return nil, exc
		}
		return nil, exception.InternalServerError.Append(err.Error())
	}
	return result.([]string), nil
}

func permissionRes(permission model.Permission) dto.PermissionResDto {
	return dto.PermissionResDto{
		ID: permission.ID, Code: permission.Code, Name: permission.Name,
		Description: permission.Description, Enabled: permission.Enabled,
	}
}

func roleToDto(role *model.UserRole) *dto.UserRoleResDto {
	permissions := make([]dto.PermissionResDto, len(role.Permissions))
	for i, permission := range role.Permissions {
		permissions[i] = permissionRes(permission)
	}
	createdAt, updatedAt := "", ""
	if role.CreatedAt != nil {
		createdAt = role.CreatedAt.Format(time.RFC3339)
	}
	if role.UpdatedAt != nil {
		updatedAt = role.UpdatedAt.Format(time.RFC3339)
	}
	return &dto.UserRoleResDto{
		ID: role.ID, TenantID: role.TenantID, Code: role.Code, Name: role.Name,
		Description: role.Description, BuiltIn: role.BuiltIn, Enabled: role.Enabled,
		Permissions: permissions, CreatedAt: createdAt, UpdatedAt: updatedAt,
	}
}

func (s *UserRoleServiceImpl) scopedRole(ctx context.Context, id uint64, preload bool) (*model.UserRole, *exception.Exception) {
	opts := []repo.QueryOption{repo.Where("id = ? AND (tenant_id = 0 OR tenant_id = ?)", id, cctx.GetTenantID(ctx))}
	if preload {
		opts = append(opts, repo.Preload("Permissions"))
	}
	role, err := s.repo.UserRole().GetOne(ctx, opts...)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, exception.InternalServerError.Append(err.Error())
	}
	if role == nil {
		return nil, exception.UserRoleNotFound
	}
	return role, nil
}

func (s *UserRoleServiceImpl) GetByID(ctx context.Context, id uint64) (*dto.UserRoleResDto, *exception.Exception) {
	role, exc := s.scopedRole(ctx, id, true)
	if exc != nil {
		return nil, exc
	}
	return roleToDto(role), nil
}

func (s *UserRoleServiceImpl) GetTable(ctx context.Context, params dto.UserRoleTableQueryReqDto) (*dto.PaginationResDto[[]*dto.UserRoleResDto], *exception.Exception) {
	roles, total, err := s.repo.UserRole().GetTable(ctx, params.Page, params.PageSize,
		repo.Where("tenant_id = 0 OR tenant_id = ?", cctx.GetTenantID(ctx)),
		repo.Preload("Permissions"), repo.Order("built_in DESC, code ASC"),
		repo.WherePtrNonEmpty("code = ?", params.Code), repo.WherePtrNonEmpty("enabled = ?", params.Enabled))
	if err != nil {
		return nil, exception.InternalServerError.Append(err.Error())
	}
	res := make([]*dto.UserRoleResDto, len(roles))
	for i, role := range roles {
		res[i] = roleToDto(role)
	}
	return utils.AssemblePaginationResDto(res, total, params.Page, params.PageSize), nil
}

func (s *UserRoleServiceImpl) validatePermissionIDs(ctx context.Context, ids []uint64) *exception.Exception {
	if len(ids) == 0 {
		return nil
	}
	permissions, err := s.repo.Permission().GetByIDs(ctx, ids, repo.Where("enabled = ?", true))
	if err != nil {
		return exception.InternalServerError.Append(err.Error())
	}
	if len(permissions) != len(ids) {
		return exception.InvalidParam.Append("one or more permissions do not exist or are disabled")
	}
	for _, permission := range permissions {
		if constant.IsTenantManagementPermission(permission.Code) {
			return exception.Forbidden.Append("tenant management permissions cannot be assigned to custom roles")
		}
	}
	actor := cctx.GetUserUniCodeFromContext(ctx)
	if actor == "" || s.access == nil {
		return exception.Forbidden.Append("authorization identity not found")
	}
	actorPermissions, exc := s.access.GetCachedPermissionCodesByUniCode(ctx, actor)
	if exc != nil {
		return exc
	}
	allowed := make(map[string]struct{}, len(actorPermissions))
	for _, code := range actorPermissions {
		allowed[code] = struct{}{}
	}
	for _, permission := range permissions {
		if _, ok := allowed[permission.Code]; !ok {
			return exception.Forbidden.Append("cannot grant permission not held by current user: " + permission.Code)
		}
	}
	return nil
}

func (s *UserRoleServiceImpl) Create(ctx context.Context, params dto.UserRoleCreateReqDto) (*dto.UserRoleResDto, *exception.Exception) {
	if !roleCodePattern.MatchString(params.Code) {
		return nil, exception.InvalidParam.Append("role code must match ^[a-z][a-z0-9_]{1,49}$")
	}
	if enum.RoleCode(params.Code).IsValid() {
		return nil, exception.Forbidden.Append("built-in role codes are reserved")
	}
	tenantID := cctx.GetTenantID(ctx)
	existing, err := s.repo.UserRole().GetOne(ctx, repo.Where("tenant_id = ? AND code = ?", tenantID, params.Code))
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, exception.InternalServerError.Append(err.Error())
	}
	if existing != nil {
		return nil, exception.UserRoleAlreadyExists
	}
	if exc := s.validatePermissionIDs(ctx, params.PermissionIDs); exc != nil {
		return nil, exc
	}
	enabled := true
	if params.Enabled != nil {
		enabled = *params.Enabled
	}
	role := &model.UserRole{TenantID: tenantID, Code: params.Code, Name: params.Name, Description: params.Description, Enabled: enabled}
	if err := s.repo.Transaction(ctx, func(tx *gorm.DB) error {
		roleRepo := s.repo.UserRole().WithTx(tx)
		if err := roleRepo.Create(ctx, role); err != nil {
			return err
		}
		return roleRepo.ReplaceRolePermissions(ctx, role.ID, params.PermissionIDs)
	}); err != nil {
		return nil, exception.InternalServerError.Append(err.Error())
	}
	s.InvalidateTenantAccessCaches(ctx, tenantID)
	return s.GetByID(ctx, role.ID)
}

func (s *UserRoleServiceImpl) Update(ctx context.Context, id uint64, params dto.UserRoleUpdateReqDto) (*dto.UserRoleResDto, *exception.Exception) {
	role, exc := s.scopedRole(ctx, id, false)
	if exc != nil {
		return nil, exc
	}
	if role.BuiltIn || role.TenantID == 0 {
		return nil, exception.Forbidden.Append("built-in roles cannot be modified")
	}
	updates := map[string]any{}
	if params.Code != nil {
		if !roleCodePattern.MatchString(*params.Code) {
			return nil, exception.InvalidParam.Append("invalid role code")
		}
		if enum.RoleCode(*params.Code).IsValid() {
			return nil, exception.Forbidden.Append("built-in role codes are reserved")
		}
		existing, err := s.repo.UserRole().GetOne(ctx,
			repo.Where("tenant_id = ? AND code = ? AND id <> ?", cctx.GetTenantID(ctx), *params.Code, id))
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, exception.InternalServerError.Append(err.Error())
		}
		if existing != nil {
			return nil, exception.UserRoleAlreadyExists
		}
		updates["code"] = *params.Code
	}
	if params.Name != nil {
		updates["name"] = *params.Name
	}
	if params.Description != nil {
		updates["description"] = *params.Description
	}
	if params.Enabled != nil {
		updates["enabled"] = *params.Enabled
	}
	if len(updates) > 0 {
		rows, err := s.repo.UserRole().UpdateByMapWithTenant(ctx, id, cctx.GetTenantID(ctx), updates)
		if err != nil {
			return nil, exception.InternalServerError.Append(err.Error())
		}
		if rows == 0 {
			return nil, exception.UserRoleNotFound
		}
	}
	s.InvalidateTenantAccessCaches(ctx, cctx.GetTenantID(ctx))
	return s.GetByID(ctx, id)
}

func (s *UserRoleServiceImpl) SetPermissions(ctx context.Context, id uint64, params dto.UserRoleSetPermissionsReqDto) (*dto.UserRoleResDto, *exception.Exception) {
	role, exc := s.scopedRole(ctx, id, false)
	if exc != nil {
		return nil, exc
	}
	if role.BuiltIn || role.TenantID == 0 {
		return nil, exception.Forbidden.Append("built-in role permissions cannot be modified")
	}
	if exc := s.validatePermissionIDs(ctx, params.PermissionIDs); exc != nil {
		return nil, exc
	}
	if err := s.repo.UserRole().ReplaceRolePermissions(ctx, role.ID, params.PermissionIDs); err != nil {
		return nil, exception.InternalServerError.Append(err.Error())
	}
	s.InvalidateTenantAccessCaches(ctx, cctx.GetTenantID(ctx))
	return s.GetByID(ctx, id)
}

func (s *UserRoleServiceImpl) Delete(ctx context.Context, id uint64) *exception.Exception {
	role, exc := s.scopedRole(ctx, id, false)
	if exc != nil {
		return exc
	}
	if role.BuiltIn || role.TenantID == 0 {
		return exception.Forbidden.Append("built-in roles cannot be deleted")
	}
	rows, err := s.repo.UserRole().HardDeleteWithTenant(ctx, id, cctx.GetTenantID(ctx))
	if err != nil {
		return exception.InternalServerError.Append(err.Error())
	}
	if rows == 0 {
		return exception.UserRoleNotFound
	}
	s.InvalidateTenantAccessCaches(ctx, cctx.GetTenantID(ctx))
	return nil
}

func (s *UserRoleServiceImpl) InvalidateTenantAccessCaches(ctx context.Context, tenantID uint64) {
	if s.access != nil {
		s.access.InvalidateTenantAccessCaches(ctx, tenantID)
	}
}

func (s *UserRoleServiceImpl) DeleteAccessCache(ctx context.Context, uniCode string) {
	if s.access != nil {
		s.access.DeleteAccessCache(ctx, uniCode)
	}
}

func (s *UserRoleServiceImpl) DeleteTenantAccessCache(ctx context.Context, tenantID uint64, uniCode string) {
	if s.access != nil {
		s.access.DeleteTenantAccessCache(ctx, tenantID, uniCode)
	}
}
