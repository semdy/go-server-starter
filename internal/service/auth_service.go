package service

import (
	"context"
	"errors"
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
}

type AuthServiceImpl struct {
	repo   repo.Repo
	jwt    *jwt.JWT
	taskq  *taskq.Client
	logger *zap.Logger
}

func NewAuthService(repo repo.Repo, jwt *jwt.JWT, taskq *taskq.Client, logger *zap.Logger) AuthService {
	return &AuthServiceImpl{
		repo:   repo,
		jwt:    jwt,
		taskq:  taskq,
		logger: logger,
	}
}

// loginOrRegister is a shared helper that looks up a user by a query condition,
// and if not found, creates a new user with the given fields and binds the "user" role.
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

	if user == nil {
		tx := s.repo.DB().Begin()
		userRepo := s.repo.User().WithTx(tx)
		userRoleRepo := s.repo.UserRole().WithTx(tx)

		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()

		uniCode, err := userRepo.GenerateUniCode(ctx)
		if err != nil {
			tx.Rollback()
			return nil, exception.InternalServerError.Append(err.Error())
		}

		// Bind default "user" role
		role, err := userRoleRepo.GetOne(ctx, repo.Where("code = ?", enum.RoleCodeUser))
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			return nil, exception.InternalServerError.Append(err.Error())
		}
		if role == nil {
			tx.Rollback()
			return nil, exception.UserRoleNotFound
		}

		user = newUser(uniCode)
		user.Roles = []model.UserRole{*role}

		if err := userRepo.Create(ctx, user); err != nil {
			tx.Rollback()
			return nil, exception.InternalServerError.Append(err.Error())
		}
		if err := tx.Commit().Error; err != nil {
			return nil, exception.InternalServerError.Append(err.Error())
		}

		// Enqueue welcome email for new user (idempotent, fire-and-forget)
		if s.taskq != nil && user.Email != "" {
			task, _ := taskq.NewEmailWelcomeTask(taskq.EmailWelcomePayload{
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

	token, err := s.jwt.GenerateToken(user.UniCode, deviceType)
	if err != nil {
		return nil, exception.InternalServerError.Append(err.Error())
	}

	return &dto.AuthTokenResDto{Token: token}, nil
}

func (s *AuthServiceImpl) LoginByMobileAndCode(ctx context.Context, deviceType enum.DeviceType, params dto.AuthLoginByMobileAndCodeReqDto) (*dto.AuthTokenResDto, *exception.Exception) {
	// TODO: Implement real SMS verification code validation
	if err := verifyCode(params.Code, exception.UserMobileVerificationCodeIsIncorrect); err != nil {
		return nil, err
	}

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
	// TODO: Implement real email verification code validation
	if err := verifyCode(params.Code, exception.UserEmailVerificationCodeIsIncorrect); err != nil {
		return nil, err
	}

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

// verifyCode validates a verification code is not empty, returning the given exception if invalid.
// TODO: Replace with real SMS/email verification service.
func verifyCode(code string, invalidErr *exception.Exception) *exception.Exception {
	if code == "" {
		return invalidErr
	}
	// TODO: Call external verification service to validate the code
	return nil
}
