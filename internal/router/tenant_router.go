package router

import "go-server-starter/internal/enum"

func (r *Router) SetupTenantRoutes() {
	router := r.router.Group("/admin/tenants")
	router.Use(r.jwt.JWT(), r.auth.RoleCheckAny(enum.RoleCodeSuperAdmin))
	router.Use(r.ratelimit.RateLimitByUser(60, "ADMIN-TENANT"))
	{
		router.GET("/code", r.handler.Tenant().GenerateCode)
		router.GET("/:id", r.handler.Tenant().GetByID)
		router.GET("", r.handler.Tenant().GetTable)
		router.POST("", r.handler.Tenant().Create)
		router.PUT("/:id", r.handler.Tenant().Update)
		router.DELETE("/:id", r.handler.Tenant().Delete)
	}
}
