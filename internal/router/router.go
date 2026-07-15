package router

import (
	"go-server-starter/internal/handler"
	"go-server-starter/internal/middleware"
	"go-server-starter/pkg/auth"
	"go-server-starter/pkg/jwt"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "go-server-starter/docs" // swag generated docs
)

type Router struct {
	handler   handler.Handler
	router    *gin.RouterGroup
	jwt       *jwt.JWT
	auth      auth.Auth
	ratelimit *middleware.RateLimit
}

func NewRouter(handler handler.Handler, router *gin.RouterGroup, jwt *jwt.JWT, auth auth.Auth, ratelimit *middleware.RateLimit) *Router {
	return &Router{handler: handler, router: router, jwt: jwt, auth: auth, ratelimit: ratelimit}
}

func (r *Router) SetupRoutes() {
	// Hello (公开接口)
	r.router.GET("/hello", r.handler.Hello().Hello)

	// Healthz (存活+就绪检查)
	r.router.GET("/healthz", r.handler.Health().Healthz)

	// Auth - 认证相关
	r.SetupAuthRoutes()

	// User - 用户相关
	r.SetupUserRoutes()

	// UserRole - 角色管理（管理员）
	r.SetupUserRoleRoutes()

	// Permission - 权限管理
	r.SetupPermissionRoutes()

	// Admin DeadLetters - 死信管理（MySQL 落库 + 重试）
	r.SetupDeadLetterRoutes()

	// Admin Tenants - 租户管理
	r.SetupTenantRoutes()

	// Swagger API 文档
	r.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler,
		func(c *ginSwagger.Config) { c.PersistAuthorization = true },
	))
}
