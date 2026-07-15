package router

import (
	"go-server-starter/internal/constant"
)

func (r *Router) SetupUserRoutes() {
	router := r.router.Group("/user")
	router.Use(r.jwt.JWT())
	{
		// My info — user-based, 60/min
		my := router.Group("")
		my.Use(r.ratelimit.RateLimitByUser(60, "USER"))
		{
			my.GET("/my-info", r.handler.User().GetMyInfo)
			my.PUT("/my-info", r.handler.User().UpdateMyInfo)
		}

		// Admin: user management — user-based, 120/min
		admin := router.Group("/admin")
		admin.Use(r.ratelimit.RateLimitByUser(120, "ADMIN-USER"))
		{
			admin.GET("/table", r.auth.PermissionCheckAny(constant.PermissionUserRead), r.handler.User().GetTable)
			admin.GET("/:id", r.auth.PermissionCheckAny(constant.PermissionUserRead), r.handler.User().GetInfoByID)
			admin.POST("", r.auth.PermissionCheckAny(constant.PermissionUserCreate), r.handler.User().UserCreate)
			admin.PUT("/:id", r.auth.PermissionCheckAny(constant.PermissionUserUpdate), r.handler.User().UserUpdate)
			admin.PUT("/:id/roles", r.auth.PermissionCheckAny(constant.PermissionUserAssignRoles), r.handler.User().SetRoles)
			admin.DELETE("/:id", r.auth.PermissionCheckAny(constant.PermissionUserDelete), r.handler.User().UserDelete)
		}
	}
}
