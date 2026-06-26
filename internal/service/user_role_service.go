package service

import (
	"encoding/json"
	"errors"
	"go-server-starter/internal/constant"
	"go-server-starter/internal/ctx"
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
	"gorm.io/gorm"
)

type UserRoleService interface {
	GetRolesCodeByUniCode(ctx *ctx.Context, uniCode string) ([]enum.RoleCode, *exception.Exception)
	GetCachedRolesCodeByUniCode(ctx *ctx.Context, uniCode string) ([]enum.RoleCode, *exception.Exception)
	GetByID(ctx *ctx.Context, id uint64) (*dto.UserRoleResDto, *exception.Exception)
	GetTable(ctx *ctx.Context, params dto.UserRoleTableQueryReqDto) (*dto.PaginationResDto[[]*dto.UserRoleResDto], *exception.Exception)
	Create(ctx *ctx.Context, params dto.UserRoleCreateReqDto) (*dto.UserRoleResDto, *exception.Exception)
	Update(ctx *ctx.Context, id uint64, params dto.UserRoleUpdateReqDto) (*dto.UserRoleResDto, *exception.Exception)
	Delete(ctx *ctx.Context, id uint64) *exception.Exception
}

type UserRoleServiceImpl struct {
	repo   repo.Repo
	redis  *redis.Client
	logger *zap.Logger
}

func NewUserRoleService(repo repo.Repo, redis *redis.Client, logger *zap.Logger) UserRoleService {
	return &UserRoleServiceImpl{
		repo:   repo,
		redis:  redis,
		logger: logger,
	}
}

func (s *UserRoleServiceImpl) GetRolesCodeByUniCode(ctx *ctx.Context, uniCode string) ([]enum.RoleCode, *exception.Exception) {
	roles, err := s.repo.User().GetRolesByUniCode(ctx.Ctx, uniCode)
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

func (s *UserRoleServiceImpl) GetCachedRolesCodeByUniCode(ctx *ctx.Context, uniCode string) ([]enum.RoleCode, *exception.Exception) {
	dataStr, err := s.redis.Get(ctx.Ctx, constant.RedisKeyOfAuthRoles(uniCode)).Result()
	if err != nil {
		// 如果redis 非正常报错，则返回错误
		if err != goredis.Nil {
			s.logger.Error("get cached roles code by uni code failed", zap.String("uniCode", uniCode), zap.Error(err))
			return nil, exception.InternalServerError.Append(err.Error())
		} else {
			// 如果redis 正常报错（goredis.Nil），则获取数据库中的角色
			roles, exc := s.GetRolesCodeByUniCode(ctx, uniCode)
			if exc != nil {
				return nil, exc
			}
			// 将角色转换为JSON
			rolesJSON, err := json.Marshal(roles)
			if err != nil {
				s.logger.Error("marshal roles code by uni code failed", zap.String("uniCode", uniCode), zap.Error(err))
				return nil, exception.InternalServerError.Append(err.Error())
			}
			// 将角色缓存到redis
			if err := s.redis.Set(ctx.Ctx, constant.RedisKeyOfAuthRoles(uniCode), rolesJSON, constant.REDIS_EXPIRE_OF_AUTH_ROLES).Err(); err != nil {
				s.logger.Error("set cached roles code by uni code failed", zap.String("uniCode", uniCode), zap.Error(err))
				return nil, exception.InternalServerError.Append(err.Error())
			}
			return roles, nil
		}
	} else {
		// 如果redis 正常返回，则将dataStr 反序列化为角色
		var roles []enum.RoleCode
		if err := json.Unmarshal([]byte(dataStr), &roles); err != nil {
			s.logger.Error("unmarshal roles code by uni code failed", zap.String("uniCode", uniCode), zap.Error(err))
			return nil, exception.InternalServerError.Append(err.Error())
		}
		return roles, nil
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

func (s *UserRoleServiceImpl) GetByID(ctx *ctx.Context, id uint64) (*dto.UserRoleResDto, *exception.Exception) {
	role, err := s.repo.UserRole().GetByID(ctx.Ctx, id)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, exception.InternalServerError.Append(err.Error())
	}
	if role == nil {
		return nil, exception.UserRoleNotFound
	}
	return toUserRoleResDto(role), nil
}

func (s *UserRoleServiceImpl) GetTable(ctx *ctx.Context, params dto.UserRoleTableQueryReqDto) (*dto.PaginationResDto[[]*dto.UserRoleResDto], *exception.Exception) {
	opts := []repo.QueryOption{
		repo.Order("id ASC"),
		repo.WherePtrNonEmpty("code = ?", params.Code),
		repo.WherePtrNonEmpty("enabled = ?", params.Enabled),
	}
	roles, total, err := s.repo.UserRole().GetTable(ctx.Ctx, params.Page, params.PageSize, opts...)
	if err != nil {
		return nil, exception.InternalServerError.Append(err.Error())
	}
	res := make([]*dto.UserRoleResDto, len(roles))
	for i, role := range roles {
		res[i] = toUserRoleResDto(role)
	}
	return utils.AssemblePaginationResDto(res, total, params.Page, params.PageSize), nil
}

func (s *UserRoleServiceImpl) Create(ctx *ctx.Context, params dto.UserRoleCreateReqDto) (*dto.UserRoleResDto, *exception.Exception) {
	code, err := enum.ParseRoleCode(params.Code)
	if err != nil {
		return nil, exception.InvalidParam.Append("invalid role code: " + params.Code)
	}

	// Check if role already exists
	existing, err := s.repo.UserRole().GetOne(ctx.Ctx, repo.Where("code = ?", code))
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
	if err := s.repo.UserRole().Create(ctx.Ctx, role); err != nil {
		return nil, exception.InternalServerError.Append(err.Error())
	}
	return toUserRoleResDto(role), nil
}

func (s *UserRoleServiceImpl) Update(ctx *ctx.Context, id uint64, params dto.UserRoleUpdateReqDto) (*dto.UserRoleResDto, *exception.Exception) {
	role, err := s.repo.UserRole().GetByID(ctx.Ctx, id)
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

	if err := s.repo.UserRole().UpdateByZeroFields(ctx.Ctx, id, role); err != nil {
		return nil, exception.InternalServerError.Append(err.Error())
	}
	return toUserRoleResDto(role), nil
}

func (s *UserRoleServiceImpl) Delete(ctx *ctx.Context, id uint64) *exception.Exception {
	if err := s.repo.UserRole().SoftDelete(ctx.Ctx, id); err != nil {
		return nil
	}
	return nil
}
