package router

import "go-server-starter/internal/constant"

func (r *Router) SetupTenantRoutes() {
	router := r.router.Group("/admin/tenants")
	router.Use(r.jwt.JWT())
	router.Use(r.ratelimit.RateLimitByUser(60, "ADMIN-TENANT"))
	{
		router.GET("/code", r.auth.PermissionCheckAny(constant.PermissionTenantCreate), r.handler.Tenant().GenerateCode)
		router.GET("/:id", r.auth.PermissionCheckAny(constant.PermissionTenantRead), r.handler.Tenant().GetByID)
		router.GET("", r.auth.PermissionCheckAny(constant.PermissionTenantRead), r.handler.Tenant().GetTable)
		router.POST("", r.auth.PermissionCheckAny(constant.PermissionTenantCreate), r.handler.Tenant().Create)
		router.PUT("/:id", r.auth.PermissionCheckAny(constant.PermissionTenantUpdate), r.handler.Tenant().Update)
		router.DELETE("/:id", r.auth.PermissionCheckAny(constant.PermissionTenantDelete), r.handler.Tenant().Delete)
	}
}
