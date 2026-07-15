package service

import (
	"context"
	"errors"
	cctx "go-server-starter/internal/ctx"
	"go-server-starter/internal/dto"
	"go-server-starter/internal/enum"
	"go-server-starter/internal/exception"
	"go-server-starter/internal/model"
	"go-server-starter/internal/repo"
	"go-server-starter/pkg/jwt"
	"go-server-starter/pkg/taskq"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AuthService interface {
	LoginByMobileAndCode(ctx context.Context, deviceType enum.DeviceType, params dto.AuthLoginByMobileAndCodeReqDto) (*dto.AuthTokenResDto, *exception.Exception)
	LoginByEmailAndCode(ctx context.Context, deviceType enum.DeviceType, params dto.AuthLoginByEmailAndCodeReqDto) (*dto.AuthTokenResDto, *exception.Exception)
	SwitchTenant(ctx context.Context, uniCode string, deviceType enum.DeviceType, params dto.SwitchTenantReqDto) (*dto.AuthTokenResDto, *exception.Exception)
}

type AuthServiceImpl struct {
	repo   repo.Repo
	jwt    *jwt.JWT
	taskq  *taskq.Client
	logger *zap.Logger
	access PermissionService
}

func NewAuthService(repo repo.Repo, jwt *jwt.JWT, access PermissionService, taskq *taskq.Client, logger *zap.Logger) AuthService {
	return &AuthServiceImpl{repo: repo, jwt: jwt, access: access, taskq: taskq, logger: logger}
}

// loginOrRegister looks up a user, returns their tenant_id from DB, or auto-generates one for new users.
func (s *AuthServiceImpl) loginOrRegister(
	ctx context.Context,
	deviceType enum.DeviceType,
	lookupOpt repo.QueryOption,
	newUser func(uniCode string) *model.User,
) (*dto.AuthTokenResDto, *exception.Exception) {
	user, err := s.repo.User().GetOne(ctx, lookupOpt)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, exception.InternalServerError.Append(err.Error())
	}

	if user != nil {
		if !user.Active {
			return nil, exception.Forbidden.Append("user is disabled")
		}
		// Verify tenant is active
		tenant, tenantErr := s.repo.Tenant().GetByID(ctx, user.TenantID)
		if tenantErr != nil || tenant == nil || !tenant.Active {
			return nil, exception.Forbidden.Append("tenant is disabled or deleted")
		}
	}

	if user == nil {
		tx := s.repo.DB().Begin()
		userRepo := s.repo.User().WithTx(tx)
		userRoleRepo := s.repo.UserRole().WithTx(tx)

		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
				s.logger.Error("panic in loginOrRegister", zap.Any("panic", r))
				panic(r) // re-panic after cleanup
			}
		}()

		uniCode, err := userRepo.GenerateUniCode(ctx)
		if err != nil {
			tx.Rollback()
			return nil, exception.InternalServerError.Append(err.Error())
		}

		// Bind default "user" role
		role, err := userRoleRepo.GetOne(ctx, repo.Where("tenant_id = 0 AND code = ?", enum.RoleCodeUser.String()))
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			return nil, exception.InternalServerError.Append(err.Error())
		}
		if role == nil {
			tx.Rollback()
			return nil, exception.UserRoleNotFound
		}

		defaultTenant, tenantErr := s.repo.Tenant().GetOne(ctx, repo.Where("code = ?", "default"))
		if tenantErr != nil || defaultTenant == nil {
			tx.Rollback()
			return nil, exception.InternalServerError.Append("default tenant not found")
		}

		user = newUser(uniCode)
		user.TenantID = defaultTenant.ID
		user.Active = true
		if err := userRepo.Create(ctx, user); err != nil {
			tx.Rollback()
			return nil, exception.InternalServerError.Append(err.Error())
		}
		if err := userRepo.AddTenantMembership(ctx, user.ID, defaultTenant.ID); err != nil {
			tx.Rollback()
			return nil, exception.InternalServerError.Append(err.Error())
		}
		if err := userRoleRepo.AddUserRole(ctx, user.ID, defaultTenant.ID, role.ID); err != nil {
			tx.Rollback()
			return nil, exception.InternalServerError.Append(err.Error())
		}
		if err := tx.Commit().Error; err != nil {
			return nil, exception.InternalServerError.Append(err.Error())
		}

		// Enqueue welcome email for new user (idempotent, fire-and-forget)
		if s.taskq != nil && user.Email != "" {
			task, _ := taskq.NewEmailWelcomeTask(taskq.EmailWelcomePayload{
				TenantID:    &user.TenantID,
				UserUniCode: user.UniCode,
				Email:       user.Email,
				Nickname:    user.Nickname,
			})
			if task != nil {
				uniqueKey := taskq.WelcomeEmailUniqueKey(user.UniCode)
				if _, err := s.taskq.EnqueueUnique(ctx, task, uniqueKey, 24*time.Hour, taskq.RetryByType(taskq.TaskEmailWelcome)...); err != nil {
					s.logger.Warn("failed to enqueue welcome email", zap.Error(err))
				}
			}
		}
	}

	token, err := s.jwt.GenerateToken(user.UniCode, user.TenantID, deviceType)
	if err != nil {
		return nil, exception.InternalServerError.Append(err.Error())
	}

	res := &dto.AuthTokenResDto{Token: token, Roles: []string{}, Permissions: []string{}}
	if s.access != nil {
		access, exc := s.access.GetMyAccess(cctx.WithTenant(ctx, user.TenantID), user.UniCode)
		if exc != nil {
			return nil, exc
		}
		res.Roles = access.Roles
		res.Permissions = access.Permissions
	}
	return res, nil
}

func (s *AuthServiceImpl) LoginByMobileAndCode(ctx context.Context, deviceType enum.DeviceType, params dto.AuthLoginByMobileAndCodeReqDto) (*dto.AuthTokenResDto, *exception.Exception) {
	params.Mobile = strings.ReplaceAll(params.Mobile, " ", "")

	return s.loginOrRegister(ctx, deviceType,
		repo.Where("mobile = ? AND country_code = ?", params.Mobile, params.CountryCode),
		func(uniCode string) *model.User {
			return &model.User{
				UniCode:     uniCode,
				Mobile:      params.Mobile,
				CountryCode: params.CountryCode,
				Nickname:    params.Mobile,
			}
		},
	)
}

func (s *AuthServiceImpl) LoginByEmailAndCode(ctx context.Context, deviceType enum.DeviceType, params dto.AuthLoginByEmailAndCodeReqDto) (*dto.AuthTokenResDto, *exception.Exception) {
	params.Email = strings.ToLower(strings.TrimSpace(params.Email))

	return s.loginOrRegister(ctx, deviceType,
		repo.Where("email = ?", params.Email),
		func(uniCode string) *model.User {
			return &model.User{
				UniCode:  uniCode,
				Email:    params.Email,
				Nickname: params.Email,
			}
		},
	)
}

func (s *AuthServiceImpl) SwitchTenant(ctx context.Context, uniCode string, deviceType enum.DeviceType, params dto.SwitchTenantReqDto) (*dto.AuthTokenResDto, *exception.Exception) {
	user, err := s.repo.User().GetByUniCode(ctx, uniCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, exception.UserNotFound
		}
		return nil, exception.InternalServerError.Append(err.Error())
	}

	tenant, tenantErr := s.repo.Tenant().GetByID(ctx, params.TenantID)
	if tenantErr != nil || tenant == nil || !tenant.Active {
		return nil, exception.Forbidden.Append("tenant not found or disabled")
	}

	isMember := user.TenantID == tenant.ID
	if !isMember {
		var memberErr error
		isMember, memberErr = s.repo.User().HasTenantMembership(ctx, user.ID, tenant.ID)
		if memberErr != nil {
			return nil, exception.InternalServerError.Append(memberErr.Error())
		}
	}
	if !isMember {
		return nil, exception.Forbidden.Append("user is not a member of this tenant")
	}

	token, tokenErr := s.jwt.GenerateToken(user.UniCode, tenant.ID, deviceType)
	if tokenErr != nil {
		return nil, exception.InternalServerError.Append(tokenErr.Error())
	}
	res := &dto.AuthTokenResDto{Token: token, Roles: []string{}, Permissions: []string{}}
	if s.access != nil {
		access, exc := s.access.GetMyAccess(cctx.WithTenant(ctx, tenant.ID), user.UniCode)
		if exc != nil {
			return nil, exc
		}
		res.Roles = access.Roles
		res.Permissions = access.Permissions
	}
	return res, nil
}
