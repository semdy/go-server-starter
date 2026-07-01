package router

func (r *Router) SetupAuthRoutes() {
	router := r.router.Group("/auth")
	router.Use(r.ratelimit.RateLimit(10, "AUTH")) // IP-based, 10/min
	{
		router.POST("/login/mobile", r.handler.Auth().LoginByMobileAndCode)
		router.POST("/login/email", r.handler.Auth().LoginByEmailAndCode)
		router.POST("/send-sms-code", r.handler.Auth().SendSmsCode)
		router.POST("/send-email-code", r.handler.Auth().SendEmailCode)
		router.POST("/switch-tenant", r.jwt.JWT(), r.handler.Auth().SwitchTenant)
	}
}
