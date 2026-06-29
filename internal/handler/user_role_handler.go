package handler

import (
	"go-server-starter/internal/ctx"
	"go-server-starter/internal/dto"
	"go-server-starter/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UserRoleHandler interface {
	GetByID(c *gin.Context)
	GetTable(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
}

type UserRoleHandlerImpl struct {
	logger  *zap.Logger
	service service.Service
}

func NewUserRoleHandler(logger *zap.Logger, service service.Service) UserRoleHandler {
	return &UserRoleHandlerImpl{logger: logger, service: service}
}

func (h *UserRoleHandlerImpl) GetByID(c *gin.Context) {
	var appCtx = ctx.FromGinCtx(c)
	id, err := appCtx.GetPathParamID("id")
	if err != nil {
		appCtx.ToError(err)
		return
	}
	res, err := h.service.UserRole().GetByID(appCtx.Ctx, id)
	if err != nil {
		appCtx.ToError(err)
		return
	}
	appCtx.ToSuccess(res)
}

func (h *UserRoleHandlerImpl) GetTable(c *gin.Context) {
	var appCtx = ctx.FromGinCtx(c)
	var params dto.UserRoleTableQueryReqDto
	if err := appCtx.ShouldBind(&params); err != nil {
		appCtx.ToError(err)
		return
	}
	res, err := h.service.UserRole().GetTable(appCtx.Ctx, params)
	if err != nil {
		appCtx.ToError(err)
		return
	}
	appCtx.ToSuccess(res)
}

func (h *UserRoleHandlerImpl) Create(c *gin.Context) {
	var appCtx = ctx.FromGinCtx(c)
	var params dto.UserRoleCreateReqDto
	if err := appCtx.ShouldBind(&params); err != nil {
		appCtx.ToError(err)
		return
	}
	res, err := h.service.UserRole().Create(appCtx.Ctx, params)
	if err != nil {
		appCtx.ToError(err)
		return
	}
	appCtx.ToSuccess(res)
}

func (h *UserRoleHandlerImpl) Update(c *gin.Context) {
	var appCtx = ctx.FromGinCtx(c)
	id, err := appCtx.GetPathParamID("id")
	if err != nil {
		appCtx.ToError(err)
		return
	}
	var params dto.UserRoleUpdateReqDto
	if err := appCtx.ShouldBind(&params); err != nil {
		appCtx.ToError(err)
		return
	}
	res, err := h.service.UserRole().Update(appCtx.Ctx, id, params)
	if err != nil {
		appCtx.ToError(err)
		return
	}
	appCtx.ToSuccess(res)
}

func (h *UserRoleHandlerImpl) Delete(c *gin.Context) {
	var appCtx = ctx.FromGinCtx(c)
	id, err := appCtx.GetPathParamID("id")
	if err != nil {
		appCtx.ToError(err)
		return
	}
	if exc := h.service.UserRole().Delete(appCtx.Ctx, id); exc != nil {
		appCtx.ToError(exc)
		return
	}
	appCtx.ToSuccess(nil)
}
