package router

import (
	"go-server-starter/internal/enum"
)

func (r *Router) SetupUserRoleRoutes() {
	router := r.router.Group("/role")
	router.Use(r.jwt.JWT(), r.auth.RoleCheckAny(enum.RoleCodeAdmin, enum.RoleCodeSuperAdmin))
	router.Use(r.ratelimit.RateLimitByUser(120, "ADMIN-ROLE"))
	{
		router.GET("/:id", r.handler.UserRole().GetByID)
		router.GET("/table", r.handler.UserRole().GetTable)
		router.POST("", r.handler.UserRole().Create)
		router.PUT("/:id", r.handler.UserRole().Update)
		router.DELETE("/:id", r.handler.UserRole().Delete)
	}
}
