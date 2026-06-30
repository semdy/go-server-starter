package router

import (
	"go-server-starter/internal/enum"
)

func (r *Router) SetupUserRoutes() {
	router := r.router.Group("/user")
	router.Use(r.jwt.JWT())
	{
		// My info
		router.GET("/my-info", r.handler.User().GetMyInfo)
		router.PUT("/my-info", r.handler.User().UpdateMyInfo)
		// Admin: user management
		admin := router.Group("/admin")
		admin.Use(r.auth.RoleCheckAny(enum.RoleCodeAdmin, enum.RoleCodeSuperAdmin))
		{
			admin.GET("/table", r.handler.User().GetTable)
			admin.GET("/:id", r.handler.User().GetInfoByID)
			admin.POST("", r.handler.User().UserCreate)
			admin.PUT("/:id", r.handler.User().UserUpdate)
			admin.DELETE("/:id", r.handler.User().UserDelete)
		}
	}
}
