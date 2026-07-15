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
	SetPermissions(c *gin.Context)
}

type UserRoleHandlerImpl struct {
	logger  *zap.Logger
	service service.Service
}

func NewUserRoleHandler(logger *zap.Logger, service service.Service) UserRoleHandler {
	return &UserRoleHandlerImpl{logger: logger, service: service}
}

// GetByID godoc
// @Summary      查询角色
// @Description  根据 ID 查询系统内置角色或当前租户自定义角色详情。需要 role.read 权限。
// @Tags         role
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "角色 ID"
// @Success      200  {object}  dto.UserRoleResDto
// @Failure      404  {object}  map[string]interface{}
// @Router       /role/{id} [get]
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

// GetTable godoc
// @Summary      角色列表
// @Description  分页查询系统内置角色和当前租户自定义角色。需要 role.read 权限。
// @Tags         role
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        code     query     string  false  "角色代码"
// @Param        enabled  query     bool    false  "是否启用"
// @Param        page     query     int     false  "页码"  default(1)
// @Param        pageSize query     int     false  "每页条数" default(20)
// @Success      200      {object}  dto.PaginationResDto[[]dto.UserRoleResDto]
// @Router       /role/table [get]
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

// Create godoc
// @Summary      创建角色
// @Description  在当前租户创建自定义角色，可配置当前用户持有的权限。需要 role.create 权限。
// @Tags         role
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      dto.UserRoleCreateReqDto  true  "角色信息"
// @Success      200   {object}  dto.UserRoleResDto
// @Failure      400   {object}  map[string]interface{}
// @Router       /role [post]
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

// Update godoc
// @Summary      更新角色
// @Description  更新当前租户自定义角色；内置角色不可修改。需要 role.update 权限。
// @Tags         role
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int                        true  "角色 ID"
// @Param        body  body      dto.UserRoleUpdateReqDto  true  "更新参数"
// @Success      200   {object}  dto.UserRoleResDto
// @Failure      404   {object}  map[string]interface{}
// @Router       /role/{id} [put]
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

// Delete godoc
// @Summary      删除角色
// @Description  删除当前租户自定义角色并清理授权关系；内置角色不可删除。需要 role.delete 权限。
// @Tags         role
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "角色 ID"
// @Success      200  {object}  map[string]interface{}
// @Router       /role/{id} [delete]
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

// SetPermissions godoc
// @Summary 配置角色权限
// @Description 替换当前租户自定义角色的权限；内置角色不可修改。
// @Tags role
// @Security BearerAuth
// @Param id path int true "角色 ID"
// @Param body body dto.UserRoleSetPermissionsReqDto true "权限 ID 列表"
// @Success 200 {object} dto.UserRoleResDto
// @Router /role/{id}/permissions [put]
func (h *UserRoleHandlerImpl) SetPermissions(c *gin.Context) {
	appCtx := ctx.FromGinCtx(c)
	id, exc := appCtx.GetPathParamID("id")
	if exc != nil {
		appCtx.ToError(exc)
		return
	}
	var params dto.UserRoleSetPermissionsReqDto
	if exc := appCtx.ShouldBind(&params); exc != nil {
		appCtx.ToError(exc)
		return
	}
	res, exc := h.service.UserRole().SetPermissions(appCtx.Ctx, id, params)
	if exc != nil {
		appCtx.ToError(exc)
		return
	}
	appCtx.ToSuccess(res)
}
