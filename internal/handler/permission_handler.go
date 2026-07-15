package handler

import (
	"go-server-starter/internal/ctx"
	"go-server-starter/internal/dto"
	"go-server-starter/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type PermissionHandler interface {
	GetTable(c *gin.Context)
}

type PermissionHandlerImpl struct {
	logger  *zap.Logger
	service service.Service
}

func NewPermissionHandler(logger *zap.Logger, service service.Service) PermissionHandler {
	return &PermissionHandlerImpl{logger: logger, service: service}
}

// GetTable godoc
// @Summary 权限字典
// @Description 分页查询可分配的权限代码。
// @Tags permission
// @Security BearerAuth
// @Param code query string false "权限代码"
// @Param enabled query bool false "是否启用"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} dto.PaginationResDto[[]dto.PermissionResDto]
// @Router /permission/table [get]
func (h *PermissionHandlerImpl) GetTable(c *gin.Context) {
	appCtx := ctx.FromGinCtx(c)
	var params dto.PermissionTableQueryReqDto
	if exc := appCtx.ShouldBind(&params); exc != nil {
		appCtx.ToError(exc)
		return
	}
	res, exc := h.service.Permission().GetTable(appCtx.Ctx, params)
	if exc != nil {
		appCtx.ToError(exc)
		return
	}
	appCtx.ToSuccess(res)
}
