package service

import (
	"context"
	"errors"
	"time"

	"go-server-starter/internal/dto"
	"go-server-starter/internal/exception"
	"go-server-starter/internal/model"
	"go-server-starter/internal/repo"
	"go-server-starter/pkg/snowflake"
	"go-server-starter/pkg/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type TenantService interface {
	GetByID(ctx context.Context, id uint64) (*dto.TenantResDto, *exception.Exception)
	GetTable(ctx context.Context, params dto.TenantTableQueryReqDto) (*dto.PaginationResDto[[]*dto.TenantResDto], *exception.Exception)
	Create(ctx context.Context, params dto.TenantCreateReqDto) (*dto.TenantResDto, *exception.Exception)
	Update(ctx context.Context, id uint64, params dto.TenantUpdateReqDto) (*dto.TenantResDto, *exception.Exception)
	Delete(ctx context.Context, id uint64) *exception.Exception
	// GenerateCode generates a unique tenant code using snowflake.
	GenerateCode(ctx context.Context) (string, *exception.Exception)
}

type TenantServiceImpl struct {
	repo        repo.Repo
	snowflake   *snowflake.Snowflake
	roleService UserRoleService
	logger      *zap.Logger
}

func NewTenantService(repo repo.Repo, sf *snowflake.Snowflake, roleService UserRoleService, logger *zap.Logger) TenantService {
	return &TenantServiceImpl{repo: repo, snowflake: sf, roleService: roleService, logger: logger}
}

func toTenantResDto(t *model.Tenant) *dto.TenantResDto {
	return &dto.TenantResDto{
		ID:        t.ID,
		Name:      t.Name,
		Code:      t.Code,
		Active:    t.Active,
		CreatedAt: t.CreatedAt.Format(time.RFC3339),
		UpdatedAt: t.UpdatedAt.Format(time.RFC3339),
	}
}

func (s *TenantServiceImpl) GenerateCode(ctx context.Context) (string, *exception.Exception) {
	return "t-" + s.snowflake.GenerateStringID(), nil
}

func (s *TenantServiceImpl) GetByID(ctx context.Context, id uint64) (*dto.TenantResDto, *exception.Exception) {
	t, err := s.repo.Tenant().GetByID(ctx, id)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, exception.InternalServerError.Append(err.Error())
	}
	if t == nil {
		return nil, exception.NotFound.Append("tenant not found")
	}
	return toTenantResDto(t), nil
}

func (s *TenantServiceImpl) GetTable(ctx context.Context, params dto.TenantTableQueryReqDto) (*dto.PaginationResDto[[]*dto.TenantResDto], *exception.Exception) {
	opts := []repo.QueryOption{
		repo.Order("id ASC"),
		repo.WherePtrNonEmpty("active = ?", params.Active),
	}
	entries, total, err := s.repo.Tenant().GetTable(ctx, params.Page, params.PageSize, opts...)
	if err != nil {
		return nil, exception.InternalServerError.Append(err.Error())
	}
	items := make([]*dto.TenantResDto, 0, len(entries))
	for _, e := range entries {
		items = append(items, toTenantResDto(e))
	}
	return utils.AssemblePaginationResDto(items, total, params.Page, params.PageSize), nil
}

func (s *TenantServiceImpl) Create(ctx context.Context, params dto.TenantCreateReqDto) (*dto.TenantResDto, *exception.Exception) {
	existing, err := s.repo.Tenant().GetOne(ctx, repo.Where("code = ?", params.Code))
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, exception.InternalServerError.Append(err.Error())
	}
	if existing != nil {
		return nil, exception.BadRequest.Append("tenant code already exists")
	}

	t := &model.Tenant{Name: params.Name, Code: params.Code}
	if err := s.repo.Tenant().Create(ctx, t); err != nil {
		return nil, exception.InternalServerError.Append(err.Error())
	}
	return toTenantResDto(t), nil
}

func (s *TenantServiceImpl) Update(ctx context.Context, id uint64, params dto.TenantUpdateReqDto) (*dto.TenantResDto, *exception.Exception) {
	t, err := s.repo.Tenant().GetByID(ctx, id)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, exception.InternalServerError.Append(err.Error())
	}
	if t == nil {
		return nil, exception.NotFound.Append("tenant not found")
	}
	if params.Name != nil {
		t.Name = *params.Name
	}
	if params.Active != nil {
		t.Active = *params.Active
	}
	if err := s.repo.Tenant().UpdateByZeroFields(ctx, id, t); err != nil {
		return nil, exception.InternalServerError.Append(err.Error())
	}
	// If deactivated, invalidate all role caches so affected users are blocked immediately
	if params.Active != nil && !*params.Active && s.roleService != nil {
		s.roleService.InvalidateAllRoleCaches(ctx)
	}
	return toTenantResDto(t), nil
}

func (s *TenantServiceImpl) Delete(ctx context.Context, id uint64) *exception.Exception {
	if err := s.repo.Tenant().SoftDelete(ctx, id); err != nil {
		return nil
	}
	s.roleService.InvalidateAllRoleCaches(ctx)
	return nil
}
