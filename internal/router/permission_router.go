package router

import "go-server-starter/internal/constant"

func (r *Router) SetupPermissionRoutes() {
	router := r.router.Group("/permission")
	router.Use(r.jwt.JWT(), r.auth.PermissionCheckAny(constant.PermissionPermissionRead))
	router.Use(r.ratelimit.RateLimitByUser(120, "ADMIN-PERMISSION"))
	router.GET("/table", r.handler.Permission().GetTable)
}
