package handler

import (
	"go-server-starter/internal/ctx"
	"go-server-starter/internal/dto"
	"go-server-starter/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type TenantHandler interface {
	GetByID(c *gin.Context)
	GetTable(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	GenerateCode(c *gin.Context)
}

type TenantHandlerImpl struct {
	logger  *zap.Logger
	service service.Service
}

func NewTenantHandler(logger *zap.Logger, service service.Service) TenantHandler {
	return &TenantHandlerImpl{logger: logger, service: service}
}

func (h *TenantHandlerImpl) GetByID(c *gin.Context) {
	var appCtx = ctx.FromGinCtx(c)
	id, err := appCtx.GetPathParamID("id")
	if err != nil {
		appCtx.ToError(err)
		return
	}
	res, err := h.service.Tenant().GetByID(appCtx.Ctx, id)
	if err != nil {
		appCtx.ToError(err)
		return
	}
	appCtx.ToSuccess(res)
}

func (h *TenantHandlerImpl) GetTable(c *gin.Context) {
	var appCtx = ctx.FromGinCtx(c)
	var params dto.TenantTableQueryReqDto
	if err := appCtx.ShouldBind(&params); err != nil {
		appCtx.ToError(err)
		return
	}
	res, err := h.service.Tenant().GetTable(appCtx.Ctx, params)
	if err != nil {
		appCtx.ToError(err)
		return
	}
	appCtx.ToSuccess(res)
}

func (h *TenantHandlerImpl) Create(c *gin.Context) {
	var appCtx = ctx.FromGinCtx(c)
	var params dto.TenantCreateReqDto
	if err := appCtx.ShouldBind(&params); err != nil {
		appCtx.ToError(err)
		return
	}
	res, err := h.service.Tenant().Create(appCtx.Ctx, params)
	if err != nil {
		appCtx.ToError(err)
		return
	}
	appCtx.ToSuccess(res)
}

func (h *TenantHandlerImpl) Update(c *gin.Context) {
	var appCtx = ctx.FromGinCtx(c)
	id, err := appCtx.GetPathParamID("id")
	if err != nil {
		appCtx.ToError(err)
		return
	}
	var params dto.TenantUpdateReqDto
	if err := appCtx.ShouldBind(&params); err != nil {
		appCtx.ToError(err)
		return
	}
	res, err := h.service.Tenant().Update(appCtx.Ctx, id, params)
	if err != nil {
		appCtx.ToError(err)
		return
	}
	appCtx.ToSuccess(res)
}

func (h *TenantHandlerImpl) Delete(c *gin.Context) {
	var appCtx = ctx.FromGinCtx(c)
	id, err := appCtx.GetPathParamID("id")
	if err != nil {
		appCtx.ToError(err)
		return
	}
	if err := h.service.Tenant().Delete(appCtx.Ctx, id); err != nil {
		appCtx.ToError(err)
		return
	}
	appCtx.ToSuccess(nil)
}

func (h *TenantHandlerImpl) GenerateCode(c *gin.Context) {
	var appCtx = ctx.FromGinCtx(c)
	code, err := h.service.Tenant().GenerateCode(appCtx.Ctx)
	if err != nil {
		appCtx.ToError(err)
		return
	}
	appCtx.ToSuccess(map[string]string{"code": code})
}
