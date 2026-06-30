package handler

import (
	"go-server-starter/internal/ctx"
	"go-server-starter/internal/dto"
	"go-server-starter/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UserHandler interface {
	GetMyInfo(c *gin.Context)
	UpdateMyInfo(c *gin.Context)
	GetTable(c *gin.Context)
	GetInfoByID(c *gin.Context)
	UserCreate(c *gin.Context)
	UserUpdate(c *gin.Context)
	UserDelete(c *gin.Context)
}

type UserHandlerImpl struct {
	logger  *zap.Logger
	service service.Service
}

func NewUserHandler(logger *zap.Logger, service service.Service) UserHandler {
	return &UserHandlerImpl{logger: logger, service: service}
}

// GetMyInfo godoc
// @Summary      获取当前用户信息
// @Description  根据 JWT token 返回当前登录用户的详情及角色列表
// @Tags         user
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dto.UserInfoResDto
// @Failure      401  {object}  map[string]interface{}
// @Router       /user/my-info [get]
func (h *UserHandlerImpl) GetMyInfo(c *gin.Context) {
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

// UpdateMyInfo godoc
// @Summary      更新当前用户信息
// @Description  更新当前登录用户的昵称、头像、简介
// @Tags         user
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      dto.UserUpdateInfoReqDto  true  "更新参数"
// @Success      200   {object}  dto.UserInfoResDto
// @Failure      400   {object}  map[string]interface{}
// @Router       /user/my-info [put]
func (h *UserHandlerImpl) UpdateMyInfo(c *gin.Context) {
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
	res, err := h.service.User().UpdateMyInfo(appCtx.Ctx, uniCode, params)
	if err != nil {
		appCtx.ToError(err)
		return
	}
	appCtx.ToSuccess(res)
}

// GetTable godoc
// @Summary      用户列表（管理员）
// @Description  分页查询用户列表，支持按昵称、邮箱、手机号筛选。需 admin+ 角色。
// @Tags         user
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        nickname     query     string  false  "昵称（模糊）"
// @Param        email        query     string  false  "邮箱（前缀）"
// @Param        mobile       query     string  false  "手机号（前缀）"
// @Param        countryCode  query     string  false  "国家代码（前缀）"
// @Param        page         query     int     false  "页码"  default(1)
// @Param        pageSize     query     int     false  "每页条数" default(20)
// @Success      200          {object}  dto.PaginationResDto[[]dto.UserListItemResDto]
// @Failure      401          {object}  map[string]interface{}
// @Failure      403          {object}  map[string]interface{}
// @Router       /user/admin/table [get]
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

// GetInfoByID godoc
// @Summary      查询用户（管理员）
// @Description  根据 ID 查询任意用户详情。需 admin+ 角色。
// @Tags         user
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "用户 ID"
// @Success      200  {object}  dto.UserInfoResDto
// @Failure      404  {object}  map[string]interface{}
// @Router       /user/admin/{id} [get]
func (h *UserHandlerImpl) GetInfoByID(c *gin.Context) {
	var appCtx = ctx.FromGinCtx(c)
	id, err := appCtx.GetPathParamID("id")
	if err != nil {
		appCtx.ToError(err)
		return
	}
	res, err := h.service.User().GetInfoByID(appCtx.Ctx, id)
	if err != nil {
		appCtx.ToError(err)
		return
	}
	appCtx.ToSuccess(res)
}

// UserCreate godoc
// @Summary      创建用户（管理员）
// @Description  在管理员所属租户下创建用户。同一租户内 email 不可重复。需 admin+ 角色。
// @Tags         user
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      dto.CreateUserReqDto  true  "用户信息"
// @Success      200   {object}  dto.UserInfoResDto
// @Failure      400   {object}  map[string]interface{}
// @Router       /user/admin [post]
func (h *UserHandlerImpl) UserCreate(c *gin.Context) {
	var appCtx = ctx.FromGinCtx(c)
	var params dto.CreateUserReqDto
	if err := appCtx.ShouldBind(&params); err != nil {
		appCtx.ToError(err)
		return
	}
	res, err := h.service.User().UserCreate(appCtx.Ctx, params)
	if err != nil {
		appCtx.ToError(err)
		return
	}
	appCtx.ToSuccess(res)
}

// UserUpdate godoc
// @Summary      更新用户（管理员）
// @Description  管理员更新任意用户的昵称、头像、简介。需 admin+ 角色。
// @Tags         user
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int                     true  "用户 ID"
// @Param        body  body      dto.UserUpdateInfoReqDto  true  "更新参数"
// @Success      200   {object}  dto.UserInfoResDto
// @Failure      404   {object}  map[string]interface{}
// @Router       /user/admin/{id} [put]
func (h *UserHandlerImpl) UserUpdate(c *gin.Context) {
	var appCtx = ctx.FromGinCtx(c)
	id, err := appCtx.GetPathParamID("id")
	if err != nil {
		appCtx.ToError(err)
		return
	}
	var params dto.UserUpdateInfoReqDto
	if err := appCtx.ShouldBind(&params); err != nil {
		appCtx.ToError(err)
		return
	}
	res, err := h.service.User().UserUpdate(appCtx.Ctx, id, params)
	if err != nil {
		appCtx.ToError(err)
		return
	}
	appCtx.ToSuccess(res)
}

// UserDelete godoc
// @Summary      删除用户（管理员）
// @Description  软删除用户，自动校验与管理员属于同一租户。需 admin+ 角色。
// @Tags         user
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "用户 ID"
// @Success      200  {object}  map[string]interface{}
// @Router       /user/admin/{id} [delete]
func (h *UserHandlerImpl) UserDelete(c *gin.Context) {
	var appCtx = ctx.FromGinCtx(c)
	id, err := appCtx.GetPathParamID("id")
	if err != nil {
		appCtx.ToError(err)
		return
	}
	if err := h.service.User().UserDelete(appCtx.Ctx, id); err != nil {
		appCtx.ToError(err)
		return
	}
	appCtx.ToSuccess(nil)
}
