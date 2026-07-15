package router

import "go-server-starter/internal/constant"

func (r *Router) SetupUserRoleRoutes() {
	router := r.router.Group("/role")
	router.Use(r.jwt.JWT())
	router.Use(r.ratelimit.RateLimitByUser(120, "ADMIN-ROLE"))
	{
		router.GET("/:id", r.auth.PermissionCheckAny(constant.PermissionRoleRead), r.handler.UserRole().GetByID)
		router.GET("/table", r.auth.PermissionCheckAny(constant.PermissionRoleRead), r.handler.UserRole().GetTable)
		router.POST("", r.auth.PermissionCheckAny(constant.PermissionRoleCreate), r.handler.UserRole().Create)
		router.PUT("/:id", r.auth.PermissionCheckAny(constant.PermissionRoleUpdate), r.handler.UserRole().Update)
		router.PUT("/:id/permissions", r.auth.PermissionCheckAny(constant.PermissionRoleAssignPermissions), r.handler.UserRole().SetPermissions)
		router.DELETE("/:id", r.auth.PermissionCheckAny(constant.PermissionRoleDelete), r.handler.UserRole().Delete)
	}
}
