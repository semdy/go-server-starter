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

// GetInfo godoc
// @Summary      获取当前用户信息
// @Description  根据 JWT token 中的 uniCode 返回用户详情及角色列表
// @Tags         user
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dto.UserInfoResDto
// @Failure      401  {object}  map[string]interface{}
// @Router       /user/info [get]
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

// UpdateInfo godoc
// @Summary      更新用户信息
// @Description  更新当前登录用户的昵称、头像、简介
// @Tags         user
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      dto.UserUpdateInfoReqDto  true  "更新参数"
// @Success      200   {object}  dto.UserInfoResDto
// @Failure      400   {object}  map[string]interface{}
// @Router       /user/info [put]
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

// GetTable godoc
// @Summary      用户列表（管理员）
// @Description  分页查询用户列表，支持按昵称、邮箱、手机号筛选。需要 admin 或 super_admin 角色。
// @Tags         user
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        nickname     query     string  false  "昵称（模糊）"
// @Param        email        query     string  false  "邮箱（前缀匹配）"
// @Param        mobile       query     string  false  "手机号（前缀匹配）"
// @Param        countryCode  query     string  false  "国家代码（前缀匹配）"
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
