package router

import "go-server-starter/internal/constant"

func (r *Router) SetupDeadLetterRoutes() {
	router := r.router.Group("/admin/dead-letters")
	router.Use(r.jwt.JWT())
	router.Use(r.ratelimit.RateLimitByUser(60, "ADMIN-DL"))
	{
		router.GET("", r.auth.PermissionCheckAny(constant.PermissionDeadLetterRead), r.handler.DeadLetter().List)
		router.POST("/retry", r.auth.PermissionCheckAny(constant.PermissionDeadLetterRetry), r.handler.DeadLetter().Retry)
		router.POST("/retry-all", r.auth.PermissionCheckAny(constant.PermissionDeadLetterRetry), r.handler.DeadLetter().RetryAll)
		router.DELETE("", r.auth.PermissionCheckAny(constant.PermissionDeadLetterDelete), r.handler.DeadLetter().Delete)
	}
}
