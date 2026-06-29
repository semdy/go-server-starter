package router

import "go-server-starter/internal/enum"

func (r *Router) SetupDeadLetterRoutes() {
	router := r.router.Group("/admin/dead-letters")
	router.Use(r.jwt.JWT(), r.auth.RoleCheckAny(enum.RoleCodeSuperAdmin))
	{
		router.GET("", r.handler.DeadLetter().List)
		router.POST("/retry", r.handler.DeadLetter().Retry)
		router.POST("/retry-all", r.handler.DeadLetter().RetryAll)
		router.DELETE("", r.handler.DeadLetter().Delete)
	}
}
