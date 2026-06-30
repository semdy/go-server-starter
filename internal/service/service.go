package service

import (
	"go-server-starter/internal/config"
	"go-server-starter/internal/repo"
	"go-server-starter/pkg/jwt"
	"go-server-starter/pkg/redis"
	"go-server-starter/pkg/snowflake"
	"go-server-starter/pkg/taskq"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Service interface {
	User() UserService
	UserRole() UserRoleService
	Auth() AuthService
	VerifyCode() VerifyCodeService
	DeadLetter() DeadLetterService
	Tenant() TenantService
}

type ServiceImpl struct {
	db                *gorm.DB
	config            *config.Config
	jwt               *jwt.JWT
	redis             *redis.Client
	snowflake         *snowflake.Snowflake
	logger            *zap.Logger
	userService       UserService
	userRoleService   UserRoleService
	authService       AuthService
	verifyCodeService VerifyCodeService
	deadLetterService DeadLetterService
	tenantService     TenantService
}

func NewService(db *gorm.DB, config *config.Config, jwt *jwt.JWT, redis *redis.Client, snowflake *snowflake.Snowflake, repo repo.Repo, taskqClient *taskq.Client, logger *zap.Logger) Service {
	return &ServiceImpl{
		db:                db,
		config:            config,
		jwt:               jwt,
		redis:             redis,
		snowflake:         snowflake,
		logger:            logger,
		userService:       NewUserService(repo, redis, logger),
		userRoleService:   NewUserRoleService(repo, redis, logger),
		authService:       NewAuthService(repo, jwt, taskqClient, logger),
		verifyCodeService: NewVerifyCodeService(redis.Client, taskqClient, logger),
		deadLetterService: NewDeadLetterService(repo, taskqClient, logger),
		tenantService:     NewTenantService(repo, snowflake, logger),
	}
}

func (s *ServiceImpl) User() UserService {
	return s.userService
}

func (s *ServiceImpl) UserRole() UserRoleService {
	return s.userRoleService
}

func (s *ServiceImpl) Auth() AuthService {
	return s.authService
}

func (s *ServiceImpl) VerifyCode() VerifyCodeService {
	return s.verifyCodeService
}

func (s *ServiceImpl) DeadLetter() DeadLetterService {
	return s.deadLetterService
}

func (s *ServiceImpl) Tenant() TenantService {
	return s.tenantService
}
