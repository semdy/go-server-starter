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
	SetRoles(ctx context.Context, id uint64, params dto.UserSetRolesReqDto) (*dto.UserInfoResDto, *exception.Exception)
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
	user, err := s.repo.User().GetByID(ctx, id, tenantFilter(ctx))
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, exception.InternalServerError.Append(err.Error())
	}
	if user == nil {
		return nil, exception.UserNotFound
	}
	if exc := s.loadRoles(ctx, user); exc != nil {
		return nil, exc
	}
	return user, nil
}

func (s *UserServiceImpl) GetByUniCode(ctx context.Context, uniCode string) (*model.User, *exception.Exception) {
	user, err := s.repo.User().GetOne(ctx, repo.Where("uni_code = ?", uniCode), tenantFilter(ctx))
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, exception.InternalServerError.Append(err.Error())
	}
	if user == nil {
		return nil, exception.UserNotFound
	}
	if exc := s.loadRoles(ctx, user); exc != nil {
		return nil, exc
	}
	return user, nil
}

func (s *UserServiceImpl) loadRoles(ctx context.Context, user *model.User) *exception.Exception {
	roles, err := s.repo.UserRole().GetRolesByUserAndTenant(ctx, user.ID, cctx.GetTenantID(ctx))
	if err != nil {
		return exception.InternalServerError.Append(err.Error())
	}
	user.Roles = make([]model.UserRole, len(roles))
	for i, role := range roles {
		user.Roles[i] = *role
	}
	return nil
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
		if exc := s.loadRoles(ctx, user); exc != nil {
			return nil, exc
		}
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
			roles[i] = role.Code
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
	tid := cctx.GetTenantID(ctx)
	user, err := s.repo.User().GetOne(ctx, repo.Where("uni_code = ?", uniCode), tenantFilter(ctx))
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, exception.InternalServerError.Append(err.Error())
	}
	if user == nil {
		return nil, exception.UserNotFound
	}
	if exc := s.loadRoles(ctx, user); exc != nil {
		return nil, exc
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

	updates := map[string]any{}
	if params.Nickname != nil {
		updates["nickname"] = user.Nickname
	}
	if params.AvatarURL != nil {
		updates["avatar_url"] = user.AvatarURL
	}
	if params.Desc != nil {
		updates["desc"] = user.Desc
	}
	if len(updates) == 0 {
		return userToInfoDto(user), nil
	}
	_, err = s.repo.User().UpdateByMapWithTenant(ctx, user.ID, tid, updates)
	if err != nil {
		return nil, exception.UserUpdateInfoFailed.Append(err.Error())
	}
	return userToInfoDto(user), nil
}

func userToInfoDto(user *model.User) *dto.UserInfoResDto {
	roles := make([]string, len(user.Roles))
	for i, role := range user.Roles {
		roles[i] = role.Code
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
		roleRepo := s.repo.UserRole().WithTx(tx)
		if err := userRepo.Create(ctx, user); err != nil {
			return err
		}
		if err := userRepo.AddTenantMembership(ctx, user.ID, tid); err != nil {
			return err
		}
		defaultRole, err := roleRepo.GetOne(ctx, repo.Where("tenant_id = 0 AND code = ?", "user"))
		if err != nil {
			return err
		}
		return roleRepo.AddUserRole(ctx, user.ID, tid, defaultRole.ID)
	}); err != nil {
		return nil, exception.InternalServerError.Append(err.Error())
	}
	if exc := s.loadRoles(ctx, user); exc != nil {
		return nil, exc
	}
	return userToInfoDto(user), nil
}

func (s *UserServiceImpl) UserUpdate(ctx context.Context, id uint64, params dto.UserUpdateInfoReqDto) (*dto.UserInfoResDto, *exception.Exception) {
	tid := cctx.GetTenantID(ctx)
	user, err := s.repo.User().GetByID(ctx, id)
	if err != nil || user == nil || user.TenantID != tid {
		return nil, exception.UserNotFound
	}
	if exc := s.loadRoles(ctx, user); exc != nil {
		return nil, exc
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

	updates := map[string]any{}
	if params.Nickname != nil {
		updates["nickname"] = user.Nickname
	}
	if params.AvatarURL != nil {
		updates["avatar_url"] = user.AvatarURL
	}
	if params.Desc != nil {
		updates["desc"] = user.Desc
	}
	if params.Active != nil {
		updates["active"] = user.Active
	}
	if len(updates) > 0 {
		_, err := s.repo.User().UpdateByMapWithTenant(ctx, id, tid, updates)
		if err != nil {
			return nil, exception.InternalServerError.Append(err.Error())
		}
	}
	// Invalidate role cache so disabled users take effect immediately
	if !user.Active && s.roleService != nil {
		s.roleService.DeleteAccessCache(ctx, user.UniCode)
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
	rows, err := s.repo.User().SoftDeleteWithTenant(ctx, id, tid)
	if err != nil {
		return exception.InternalServerError.Append(err.Error())
	}
	if rows == 0 {
		return exception.UserNotFound
	}
	if s.roleService != nil {
		s.roleService.DeleteAccessCache(ctx, user.UniCode)
	}
	return nil
}

func (s *UserServiceImpl) SetRoles(ctx context.Context, id uint64, params dto.UserSetRolesReqDto) (*dto.UserInfoResDto, *exception.Exception) {
	tenantID := cctx.GetTenantID(ctx)
	user, err := s.repo.User().GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, exception.UserNotFound
		}
		return nil, exception.InternalServerError.Append(err.Error())
	}
	isMember := user.TenantID == tenantID
	if !isMember {
		isMember, err = s.repo.User().HasTenantMembership(ctx, user.ID, tenantID)
	}
	if err != nil {
		return nil, exception.InternalServerError.Append(err.Error())
	}
	if !isMember {
		return nil, exception.UserNotFound
	}

	roles, err := s.repo.UserRole().GetByIDs(ctx, params.RoleIDs,
		repo.Where("enabled = ? AND (tenant_id = 0 OR tenant_id = ?)", true, tenantID))
	if err != nil {
		return nil, exception.InternalServerError.Append(err.Error())
	}
	if len(roles) != len(params.RoleIDs) {
		return nil, exception.InvalidParam.Append("one or more roles do not exist, are disabled, or belong to another tenant")
	}
	actor := cctx.GetUserUniCodeFromContext(ctx)
	actorRoles, exc := s.roleService.GetRolesCodeByUniCode(ctx, actor)
	if exc != nil {
		return nil, exc
	}
	isSuperAdmin := false
	for _, code := range actorRoles {
		if code == "super_admin" {
			isSuperAdmin = true
			break
		}
	}
	for _, role := range roles {
		if role.Code == "super_admin" && !isSuperAdmin {
			return nil, exception.Forbidden.Append("only super_admin can assign super_admin")
		}
	}
	if err := s.repo.UserRole().ReplaceUserRoles(ctx, user.ID, tenantID, params.RoleIDs); err != nil {
		return nil, exception.InternalServerError.Append(err.Error())
	}
	s.roleService.DeleteTenantAccessCache(ctx, tenantID, user.UniCode)
	if exc := s.loadRoles(ctx, user); exc != nil {
		return nil, exc
	}
	return userToInfoDto(user), nil
}
