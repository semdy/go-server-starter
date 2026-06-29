package router

func (r *Router) SetupAuthRoutes() {
	r.router.POST("/auth/login/mobile", r.handler.Auth().LoginByMobileAndCode)
	r.router.POST("/auth/login/email", r.handler.Auth().LoginByEmailAndCode)
	r.router.POST("/auth/send-sms-code", r.handler.Auth().SendSmsCode)
	r.router.POST("/auth/send-email-code", r.handler.Auth().SendEmailCode)
}
