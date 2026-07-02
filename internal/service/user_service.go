package service

import (
	"context"
	"errors"
	cctx "go-server-starter/internal/ctx"
	"go-server-starter/internal/dto"
	"go-server-starter/internal/exception"
	"go-server-starter/internal/model"
	"go-server-starter/internal/repo"
	"go-server-starter/pkg/redis"
	"go-server-starter/pkg/utils"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UserService interface {
	GetByID(ctx context.Context, id uint64) (*model.User, *exception.Exception)
	GetByUniCode(ctx context.Context, uniCode string) (*model.User, *exception.Exception)
	GetInfoByUniCode(ctx context.Context, uniCode string) (*dto.UserInfoResDto, *exception.Exception)
	UpdateMyInfo(ctx context.Context, uniCode string, params dto.UserUpdateInfoReqDto) (*dto.UserInfoResDto, *exception.Exception)
	GetTable(ctx context.Context, params dto.UserTableQueryReqDto) (*dto.PaginationResDto[[]*dto.UserListItemResDto], *exception.Exception)
	GetInfoByID(ctx context.Context, id uint64) (*dto.UserInfoResDto, *exception.Exception)
	UserCreate(ctx context.Context, params dto.CreateUserReqDto) (*dto.UserInfoResDto, *exception.Exception)
	UserUpdate(ctx context.Context, id uint64, params dto.UserUpdateInfoReqDto) (*dto.UserInfoResDto, *exception.Exception)
	UserDelete(ctx context.Context, id uint64) *exception.Exception
}

type UserServiceImpl struct {
	repo        repo.Repo
	roleService UserRoleService
	logger      *zap.Logger
}

func NewUserService(repo repo.Repo, redis *redis.Client, roleService UserRoleService, logger *zap.Logger) UserService {
	return &UserServiceImpl{
		repo:        repo,
		roleService: roleService,
		logger:      logger,
	}
}

// tenantFilter returns a Where option for the current tenant.
func tenantFilter(ctx context.Context) repo.QueryOption {
	return repo.Where("tenant_id = ?", cctx.GetTenantID(ctx))
}

func (s *UserServiceImpl) GetByID(ctx context.Context, id uint64) (*model.User, *exception.Exception) {
	user, err := s.repo.User().GetByID(ctx, id, repo.Preload("Roles"), tenantFilter(ctx))
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, exception.InternalServerError.Append(err.Error())
	}
	if user == nil {
		return nil, exception.UserNotFound
	}
	return user, nil
}

func (s *UserServiceImpl) GetByUniCode(ctx context.Context, uniCode string) (*model.User, *exception.Exception) {
	user, err := s.repo.User().GetOne(ctx, repo.Where("uni_code = ?", uniCode), repo.Preload("Roles"), tenantFilter(ctx))
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, exception.InternalServerError.Append(err.Error())
	}
	if user == nil {
		return nil, exception.UserNotFound
	}
	return user, nil
}

func (s *UserServiceImpl) GetInfoByUniCode(ctx context.Context, uniCode string) (*dto.UserInfoResDto, *exception.Exception) {
	user, err := s.GetByUniCode(ctx, uniCode)
	if err != nil {
		return nil, err
	}
	return userToInfoDto(user), nil
}

func (s *UserServiceImpl) GetTable(ctx context.Context, params dto.UserTableQueryReqDto) (*dto.PaginationResDto[[]*dto.UserListItemResDto], *exception.Exception) {
	opts := []repo.QueryOption{
		repo.Order("created_at DESC"),
		repo.Preload("Roles"),
		tenantFilter(ctx),
		repo.WhereAutoLike("nickname", params.Nickname),
		repo.WhereAutoLikePrefix("email", params.Email),
		repo.WhereAutoLikePrefix("mobile", params.Mobile),
		repo.WhereAutoLikePrefix("country_code", params.CountryCode),
	}
	users, total, err := s.repo.User().GetTable(ctx, params.Page, params.PageSize, opts...)
	if err != nil {
		return nil, exception.InternalServerError.Append(err.Error())
	}
	res := make([]*dto.UserListItemResDto, len(users))
	for i, user := range users {
		res[i] = &dto.UserListItemResDto{
			ID:          user.ID,
			CreatedAt:   user.CreatedAt.Format(time.RFC3339),
			UniCode:     user.UniCode,
			Active:      user.Active,
			Email:       user.Email,
			Mobile:      user.Mobile,
			CountryCode: user.CountryCode,
			Desc:        user.Desc,
			Nickname:    user.Nickname,
			AvatarURL:   user.AvatarURL,
		}
		roles := make([]string, len(user.Roles))
		for i, role := range user.Roles {
			roles[i] = role.Code.String()
		}
		res[i].Roles = roles
	}
	return utils.AssemblePaginationResDto(res, total, params.Page, params.PageSize), nil
}

func (s *UserServiceImpl) GetInfoByID(ctx context.Context, id uint64) (*dto.UserInfoResDto, *exception.Exception) {
	user, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return userToInfoDto(user), nil
}

func (s *UserServiceImpl) UpdateMyInfo(ctx context.Context, uniCode string, params dto.UserUpdateInfoReqDto) (*dto.UserInfoResDto, *exception.Exception) {
	user, err := s.repo.User().GetOne(ctx, repo.Where("uni_code = ?", uniCode), repo.Preload("Roles"), tenantFilter(ctx))
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, exception.InternalServerError.Append(err.Error())
	}
	if user == nil {
		return nil, exception.UserNotFound
	}

	if params.Nickname != nil {
		user.Nickname = *params.Nickname
	}
	if params.AvatarURL != nil {
		user.AvatarURL = *params.AvatarURL
	}
	if params.Desc != nil {
		user.Desc = *params.Desc
	}
	if err := s.repo.User().UpdateByZeroFields(ctx, user.ID, user); err != nil {
		return nil, exception.UserUpdateInfoFailed.Append(err.Error())
	}
	return userToInfoDto(user), nil
}

func userToInfoDto(user *model.User) *dto.UserInfoResDto {
	roles := make([]string, len(user.Roles))
	for i, role := range user.Roles {
		roles[i] = role.Code.String()
	}
	return &dto.UserInfoResDto{
		UniCode:     user.UniCode,
		Active:      user.Active,
		Email:       user.Email,
		Mobile:      user.Mobile,
		CountryCode: user.CountryCode,
		Desc:        user.Desc,
		Nickname:    user.Nickname,
		AvatarURL:   user.AvatarURL,
		Roles:       roles,
	}
}

func (s *UserServiceImpl) UserCreate(ctx context.Context, params dto.CreateUserReqDto) (*dto.UserInfoResDto, *exception.Exception) {
	tid := cctx.GetTenantID(ctx)
	existing, err := s.repo.User().GetOne(ctx, repo.Where("email = ? AND tenant_id = ?", params.Email, tid))
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, exception.InternalServerError.Append(err.Error())
	}
	if existing != nil {
		return nil, exception.BadRequest.Append("email already exists in this tenant")
	}
	uniCode, err := s.repo.User().GenerateUniCode(ctx)
	if err != nil {
		return nil, exception.InternalServerError.Append(err.Error())
	}
	user := &model.User{TenantID: tid, UniCode: uniCode, Active: true, Email: params.Email, Nickname: params.Nickname}
	if err := s.repo.Transaction(ctx, func(tx *gorm.DB) error {
		userRepo := s.repo.User().WithTx(tx)
		if err := userRepo.Create(ctx, user); err != nil {
			return err
		}
		return userRepo.AddTenantMembership(ctx, user.ID, tid)
	}); err != nil {
		return nil, exception.InternalServerError.Append(err.Error())
	}
	return userToInfoDto(user), nil
}

func (s *UserServiceImpl) UserUpdate(ctx context.Context, id uint64, params dto.UserUpdateInfoReqDto) (*dto.UserInfoResDto, *exception.Exception) {
	tid := cctx.GetTenantID(ctx)
	user, err := s.repo.User().GetByID(ctx, id, repo.Preload("Roles"))
	if err != nil || user == nil || user.TenantID != tid {
		return nil, exception.UserNotFound
	}
	if params.Nickname != nil {
		user.Nickname = *params.Nickname
	}
	if params.AvatarURL != nil {
		user.AvatarURL = *params.AvatarURL
	}
	if params.Desc != nil {
		user.Desc = *params.Desc
	}
	if params.Active != nil {
		user.Active = *params.Active
	}
	if err := s.repo.User().UpdateByZeroFields(ctx, id, user); err != nil {
		return nil, exception.InternalServerError.Append(err.Error())
	}
	// Invalidate role cache so disabled users take effect immediately
	if !user.Active && s.roleService != nil {
		s.roleService.DeleteRoleCache(ctx, user.UniCode)
	}
	return userToInfoDto(user), nil
}

func (s *UserServiceImpl) UserDelete(ctx context.Context, id uint64) *exception.Exception {
	tid := cctx.GetTenantID(ctx)
	user, err := s.repo.User().GetOne(ctx, repo.Where("id = ? AND tenant_id = ?", id, tid))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return exception.UserNotFound
		}
		return exception.InternalServerError.Append(err.Error())
	}
	if user == nil {
		return exception.UserNotFound
	}
	if err := s.repo.User().SoftDelete(ctx, id); err != nil {
		return exception.InternalServerError.Append(err.Error())
	}
	if s.roleService != nil {
		s.roleService.DeleteRoleCache(ctx, user.UniCode)
	}
	return nil
}
