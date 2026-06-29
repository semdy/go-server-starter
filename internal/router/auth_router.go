package router

func (r *Router) SetupAuthRoutes() {
	r.router.POST("/auth/login/mobile", r.handler.Auth().LoginByMobileAndCode)
	r.router.POST("/auth/login/email", r.handler.Auth().LoginByEmailAndCode)
}
