package handler

import (
	"go-server-starter/internal/ctx"
	"go-server-starter/internal/dto"
	"go-server-starter/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UserHandler interface {
	GetInfo(c *gin.Context)
	UpdateInfo(c *gin.Context)
	GetTable(c *gin.Context)
}

type UserHandlerImpl struct {
	logger  *zap.Logger
	service service.Service
}

func NewUserHandler(logger *zap.Logger, service service.Service) UserHandler {
	return &UserHandlerImpl{logger: logger, service: service}
}

func (h *UserHandlerImpl) GetInfo(c *gin.Context) {
	var appCtx = ctx.FromGinCtx(c)
	uniCode, err := appCtx.GetUserUniCode()
	if err != nil {
		appCtx.ToError(err)
		return
	}
	user, err := h.service.User().GetInfoByUniCode(appCtx.Ctx, uniCode)
	if err != nil {
		appCtx.ToError(err)
		return
	}
	appCtx.ToSuccess(user)
}

func (h *UserHandlerImpl) UpdateInfo(c *gin.Context) {
	var appCtx = ctx.FromGinCtx(c)
	var params dto.UserUpdateInfoReqDto
	if err := appCtx.ShouldBind(&params); err != nil {
		appCtx.ToError(err)
		return
	}
	uniCode, err := appCtx.GetUserUniCode()
	if err != nil {
		appCtx.ToError(err)
		return
	}
	res, err := h.service.User().UpdateInfo(appCtx.Ctx, uniCode, params)
	if err != nil {
		appCtx.ToError(err)
		return
	}
	appCtx.ToSuccess(res)
}

func (h *UserHandlerImpl) GetTable(c *gin.Context) {
	var appCtx = ctx.FromGinCtx(c)
	var params dto.UserTableQueryReqDto
	if err := appCtx.ShouldBind(&params); err != nil {
		appCtx.ToError(err)
		return
	}
	res, err := h.service.User().GetTable(appCtx.Ctx, params)
	if err != nil {
		appCtx.ToError(err)
		return
	}
	appCtx.ToSuccess(res)
}
