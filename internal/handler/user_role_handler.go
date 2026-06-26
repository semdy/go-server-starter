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
	var ctx = ctx.FromGinCtx(c)
	id, err := ctx.GetPathParamID("id")
	if err != nil {
		ctx.ToError(err)
		return
	}
	res, err := h.service.UserRole().GetByID(ctx, id)
	if err != nil {
		ctx.ToError(err)
		return
	}
	ctx.ToSuccess(res)
}

func (h *UserRoleHandlerImpl) GetTable(c *gin.Context) {
	var ctx = ctx.FromGinCtx(c)
	var params dto.UserRoleTableQueryReqDto
	if err := ctx.ShouldBind(&params); err != nil {
		ctx.ToError(err)
		return
	}
	res, err := h.service.UserRole().GetTable(ctx, params)
	if err != nil {
		ctx.ToError(err)
		return
	}
	ctx.ToSuccess(res)
}

func (h *UserRoleHandlerImpl) Create(c *gin.Context) {
	var ctx = ctx.FromGinCtx(c)
	var params dto.UserRoleCreateReqDto
	if err := ctx.ShouldBind(&params); err != nil {
		ctx.ToError(err)
		return
	}
	res, err := h.service.UserRole().Create(ctx, params)
	if err != nil {
		ctx.ToError(err)
		return
	}
	ctx.ToSuccess(res)
}

func (h *UserRoleHandlerImpl) Update(c *gin.Context) {
	var ctx = ctx.FromGinCtx(c)
	id, err := ctx.GetPathParamID("id")
	if err != nil {
		ctx.ToError(err)
		return
	}
	var params dto.UserRoleUpdateReqDto
	if err := ctx.ShouldBind(&params); err != nil {
		ctx.ToError(err)
		return
	}
	res, err := h.service.UserRole().Update(ctx, id, params)
	if err != nil {
		ctx.ToError(err)
		return
	}
	ctx.ToSuccess(res)
}

func (h *UserRoleHandlerImpl) Delete(c *gin.Context) {
	var ctx = ctx.FromGinCtx(c)
	id, err := ctx.GetPathParamID("id")
	if err != nil {
		ctx.ToError(err)
		return
	}
	if exc := h.service.UserRole().Delete(ctx, id); exc != nil {
		ctx.ToError(exc)
		return
	}
	ctx.ToSuccess(nil)
}
