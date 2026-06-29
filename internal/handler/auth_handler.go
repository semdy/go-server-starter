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
