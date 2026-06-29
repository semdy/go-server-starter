package handler

import (
	"go-server-starter/internal/ctx"
	"go-server-starter/internal/dto"
	"go-server-starter/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthHandler interface {
	LoginByMobileAndCode(c *gin.Context)
	LoginByEmailAndCode(c *gin.Context)
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
// @Description  手机号 + 验证码登录，未注册自动创建账号并绑定 user 角色。当前验证码仅校验非空，待接入真实 SMS 服务。
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.AuthLoginByMobileAndCodeReqDto  true  "登录参数"
// @Param        Device-Type  header    string  false  "设备类型: web/desktop/mobile/chromeExtension/api"
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
	res, err := h.service.Auth().LoginByMobileAndCode(appCtx.Ctx, deviceType, params)
	if err != nil {
		appCtx.ToError(err)
		return
	}
	appCtx.ToSuccess(res)
}

// LoginByEmailAndCode godoc
// @Summary      邮箱验证码登录
// @Description  邮箱 + 验证码登录，未注册自动创建账号并绑定 user 角色。当前验证码仅校验非空，待接入真实邮件服务。
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
	res, err := h.service.Auth().LoginByEmailAndCode(appCtx.Ctx, deviceType, params)
	if err != nil {
		appCtx.ToError(err)
		return
	}
	appCtx.ToSuccess(res)
}
