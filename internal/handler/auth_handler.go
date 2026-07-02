package handler

import (
	"go-server-starter/internal/ctx"
	"go-server-starter/internal/dto"
	"go-server-starter/internal/service"
	"go-server-starter/pkg/verify_code"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthHandler interface {
	LoginByMobileAndCode(c *gin.Context)
	LoginByEmailAndCode(c *gin.Context)
	SendSmsCode(c *gin.Context)
	SendEmailCode(c *gin.Context)
	SwitchTenant(c *gin.Context)
}

type AuthHandlerImpl struct {
	logger  *zap.Logger
	service service.Service
}

func NewAuthHandler(logger *zap.Logger, service service.Service) AuthHandler {
	return &AuthHandlerImpl{logger: logger, service: service}
}

// LoginByMobileAndCode godoc
// @Summary      手机验证码登录
// @Description  手机号 + 验证码登录，未注册自动创建账号并绑定 user 角色。
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.AuthLoginByMobileAndCodeReqDto  true  "登录参数"
// @Param        Device-Type  header    string  false  "设备类型: web/desktop/mobile/chromeExtension/api"
// @Param        X-Tenant-ID  header    string  false  "租户 ID (默认 default)"
// @Success      200   {object}  dto.AuthTokenResDto
// @Failure      400   {object}  map[string]interface{}
// @Router       /auth/login/mobile [post]
func (h *AuthHandlerImpl) LoginByMobileAndCode(c *gin.Context) {
	var appCtx = ctx.FromGinCtx(c)
	var params dto.AuthLoginByMobileAndCodeReqDto
	if err := appCtx.ShouldBind(&params); err != nil {
		appCtx.ToError(err)
		return
	}
	deviceType := appCtx.GetDeviceType()
	// Validate verification code
	if err := h.service.VerifyCode().Validate(verify_code.SMSTypeLogin, params.Mobile, params.Code); err != nil {
		appCtx.ToError(err)
		return
	}
	res, err := h.service.Auth().LoginByMobileAndCode(appCtx.Ctx, deviceType, params)
	if err != nil {
		appCtx.ToError(err)
		return
	}
	appCtx.ToSuccess(res)
}

// LoginByEmailAndCode godoc
// @Summary      邮箱验证码登录
// @Description  邮箱 + 验证码登录，未注册自动创建账号并绑定 user 角色。
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.AuthLoginByEmailAndCodeReqDto  true  "登录参数"
// @Param        Device-Type  header    string  false  "设备类型: web/desktop/mobile/chromeExtension/api"
// @Success      200   {object}  dto.AuthTokenResDto
// @Failure      400   {object}  map[string]interface{}
// @Router       /auth/login/email [post]
func (h *AuthHandlerImpl) LoginByEmailAndCode(c *gin.Context) {
	var appCtx = ctx.FromGinCtx(c)
	var params dto.AuthLoginByEmailAndCodeReqDto
	if err := appCtx.ShouldBind(&params); err != nil {
		appCtx.ToError(err)
		return
	}
	deviceType := appCtx.GetDeviceType()
	// Validate verification code
	if err := h.service.VerifyCode().Validate(verify_code.EmailTypeLogin, params.Email, params.Code); err != nil {
		appCtx.ToError(err)
		return
	}
	res, err := h.service.Auth().LoginByEmailAndCode(appCtx.Ctx, deviceType, params)
	if err != nil {
		appCtx.ToError(err)
		return
	}
	appCtx.ToSuccess(res)
}

// SendSmsCode godoc
// @Summary      发送短信验证码
// @Description  向指定手机号发送登录验证码。60 秒内不可重复请求。
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.SendSmsCodeReqDto  true  "手机号"
// @Success      200   {object}  map[string]interface{}
// @Router       /auth/send-sms-code [post]
func (h *AuthHandlerImpl) SendSmsCode(c *gin.Context) {
	var appCtx = ctx.FromGinCtx(c)
	var params dto.SendSmsCodeReqDto
	if err := appCtx.ShouldBind(&params); err != nil {
		appCtx.ToError(err)
		return
	}
	if err := h.service.VerifyCode().SendSmsCode(appCtx.Ctx, params); err != nil {
		appCtx.ToError(err)
		return
	}
	appCtx.ToSuccess(nil)
}

// SendEmailCode godoc
// @Summary      发送邮箱验证码
// @Description  向指定邮箱发送登录验证码。60 秒内不可重复请求。
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.SendEmailCodeReqDto  true  "邮箱"
// @Success      200   {object}  map[string]interface{}
// @Router       /auth/send-email-code [post]
func (h *AuthHandlerImpl) SendEmailCode(c *gin.Context) {
	var appCtx = ctx.FromGinCtx(c)
	var params dto.SendEmailCodeReqDto
	if err := appCtx.ShouldBind(&params); err != nil {
		appCtx.ToError(err)
		return
	}
	if err := h.service.VerifyCode().SendEmailCode(appCtx.Ctx, params); err != nil {
		appCtx.ToError(err)
		return
	}
	appCtx.ToSuccess(nil)
}

// SwitchTenant godoc
// @Summary      切换租户
// @Description  切换到指定租户，签发新的 JWT token（旧 token 仍有效）。用户必须属于目标租户（默认租户或 user_tenant_refs 成员关系）。
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      dto.SwitchTenantReqDto  true  "租户 ID"
// @Success      200   {object}  dto.AuthTokenResDto
// @Failure      403   {object}  map[string]interface{}
// @Router       /auth/switch-tenant [post]
func (h *AuthHandlerImpl) SwitchTenant(c *gin.Context) {
	var appCtx = ctx.FromGinCtx(c)
	uniCode, err := appCtx.GetUserUniCode()
	if err != nil {
		appCtx.ToError(err)
		return
	}
	var params dto.SwitchTenantReqDto
	if err := appCtx.ShouldBind(&params); err != nil {
		appCtx.ToError(err)
		return
	}
	res, err := h.service.Auth().SwitchTenant(appCtx.Ctx, uniCode, params)
	if err != nil {
		appCtx.ToError(err)
		return
	}
	appCtx.ToSuccess(res)
}
