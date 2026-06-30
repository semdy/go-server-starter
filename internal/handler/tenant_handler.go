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

// GenerateCode godoc
// @Summary      生成租户 Code
// @Description  使用 Snowflake 生成全局唯一的租户 Code。仅 super_admin 可访问。
// @Tags         tenant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]string
// @Router       /admin/tenants/code [get]
func (h *TenantHandlerImpl) GenerateCode(c *gin.Context) {
	var appCtx = ctx.FromGinCtx(c)
	code, err := h.service.Tenant().GenerateCode(appCtx.Ctx)
	if err != nil {
		appCtx.ToError(err)
		return
	}
	appCtx.ToSuccess(map[string]string{"code": code})
}

// GetByID godoc
// @Summary      查询租户
// @Description  根据 ID 查询租户详情。仅 super_admin 可访问。
// @Tags         tenant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "租户 ID"
// @Success      200  {object}  dto.TenantResDto
// @Failure      404  {object}  map[string]interface{}
// @Router       /admin/tenants/{id} [get]
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

// GetTable godoc
// @Summary      租户列表
// @Description  分页查询租户列表，支持按 status 筛选。仅 super_admin 可访问。
// @Tags         tenant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        status    query     string  false  "状态: active/disabled"
// @Param        page      query     int     false  "页码"  default(1)
// @Param        pageSize  query     int     false  "每页条数" default(20)
// @Success      200       {object}  dto.PaginationResDto[[]dto.TenantResDto]
// @Router       /admin/tenants [get]
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

// Create godoc
// @Summary      创建租户
// @Description  创建新租户。code 不可重复。仅 super_admin 可访问。
// @Tags         tenant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      dto.TenantCreateReqDto  true  "租户信息"
// @Success      200   {object}  dto.TenantResDto
// @Failure      400   {object}  map[string]interface{}
// @Router       /admin/tenants [post]
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

// Update godoc
// @Summary      更新租户
// @Description  更新租户名称或状态。仅 super_admin 可访问。
// @Tags         tenant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int                     true  "租户 ID"
// @Param        body  body      dto.TenantUpdateReqDto  true  "更新参数"
// @Success      200   {object}  dto.TenantResDto
// @Failure      404   {object}  map[string]interface{}
// @Router       /admin/tenants/{id} [put]
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

// Delete godoc
// @Summary      删除租户（软删除）
// @Description  软删除指定租户。仅 super_admin 可访问。
// @Tags         tenant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "租户 ID"
// @Success      200  {object}  map[string]interface{}
// @Router       /admin/tenants/{id} [delete]
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
