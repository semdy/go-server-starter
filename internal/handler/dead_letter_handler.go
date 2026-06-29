package handler

import (
	"go-server-starter/internal/ctx"
	"go-server-starter/internal/dto"
	"go-server-starter/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type DeadLetterHandler interface {
	List(c *gin.Context)
	Retry(c *gin.Context)
	RetryAll(c *gin.Context)
	Delete(c *gin.Context)
}

type DeadLetterHandlerImpl struct {
	logger  *zap.Logger
	service service.Service
}

func NewDeadLetterHandler(logger *zap.Logger, service service.Service) DeadLetterHandler {
	return &DeadLetterHandlerImpl{logger: logger, service: service}
}

// List godoc
// @Summary      死信列表（DB）
// @Description  分页查询已落库的死信记录。仅 super_admin 可访问。
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        taskType  query     string  false  "任务类型"
// @Param        isRetried query     bool    false  "是否已重试"
// @Param        page      query     int     false  "页码"
// @Param        pageSize  query     int     false  "每页条数"
// @Success      200       {object}  dto.PaginationResDto[[]dto.DeadLetterItem]
// @Router       /admin/dead-letters [get]
func (h *DeadLetterHandlerImpl) List(c *gin.Context) {
	var appCtx = ctx.FromGinCtx(c)
	var params dto.DeadLetterListReqDto
	if err := appCtx.ShouldBind(&params); err != nil {
		appCtx.ToError(err)
		return
	}
	res, err := h.service.DeadLetter().List(appCtx.Ctx, params)
	if err != nil {
		appCtx.ToError(err)
		return
	}
	appCtx.ToSuccess(res)
}

// Retry godoc
// @Summary      重试单条死信
// @Description  将指定死信任务重新入队，并标记为已重试。仅 super_admin 可访问。
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      dto.RetryDeadLetterReq  true  "死信 ID"
// @Success      200   {object}  map[string]interface{}
// @Router       /admin/dead-letters/retry [post]
func (h *DeadLetterHandlerImpl) Retry(c *gin.Context) {
	var appCtx = ctx.FromGinCtx(c)
	var params dto.RetryDeadLetterReq
	if err := appCtx.ShouldBind(&params); err != nil {
		appCtx.ToError(err)
		return
	}
	if err := h.service.DeadLetter().Retry(appCtx.Ctx, params.ID); err != nil {
		appCtx.ToError(err)
		return
	}
	appCtx.ToSuccess(nil)
}

// RetryAll godoc
// @Summary      批量重试死信
// @Description  将指定类型的所有未重试死信重新入队。仅 super_admin 可访问。
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        taskType  query     string  true  "任务类型"
// @Success      200       {object}  map[string]int
// @Router       /admin/dead-letters/retry-all [post]
func (h *DeadLetterHandlerImpl) RetryAll(c *gin.Context) {
	var appCtx = ctx.FromGinCtx(c)
	taskType := c.Query("taskType")
	if taskType == "" {
		appCtx.ToError(appCtx.ShouldBind(&struct{}{})) // 用已有异常
		return
	}
	count, err := h.service.DeadLetter().RetryAll(appCtx.Ctx, taskType)
	if err != nil {
		appCtx.ToError(err)
		return
	}
	appCtx.ToSuccess(map[string]int{"count": count})
}

// Delete godoc
// @Summary      删除死信记录
// @Description  物理删除指定死信记录。仅 super_admin 可访问。
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      dto.RetryDeadLetterReq  true  "死信 ID"
// @Success      200   {object}  map[string]interface{}
// @Router       /admin/dead-letters [delete]
func (h *DeadLetterHandlerImpl) Delete(c *gin.Context) {
	var appCtx = ctx.FromGinCtx(c)
	var params dto.RetryDeadLetterReq
	if err := appCtx.ShouldBind(&params); err != nil {
		appCtx.ToError(err)
		return
	}
	if err := h.service.DeadLetter().Delete(appCtx.Ctx, params.ID); err != nil {
		appCtx.ToError(err)
		return
	}
	appCtx.ToSuccess(nil)
}
