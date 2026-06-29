package router

import (
	"go-server-starter/internal/enum"
)

func (r *Router) SetupAdminTaskRoutes() {
	router := r.router.Group("/admin/tasks")
	router.Use(r.jwt.JWT(), r.auth.RoleCheckAny(enum.RoleCodeSuperAdmin))
	{
		router.GET("/archived", r.handler.Task().ListArchived)
		router.POST("/archived/run", r.handler.Task().RunArchived)
		router.POST("/archived/run-all", r.handler.Task().RunAllArchived)
		router.DELETE("/archived", r.handler.Task().DeleteArchived)
	}
}
