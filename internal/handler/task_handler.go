package handler

import (
	"go-server-starter/internal/ctx"
	"go-server-starter/internal/dto"
	"go-server-starter/internal/exception"
	"go-server-starter/pkg/taskq"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type TaskHandler interface {
	ListArchived(c *gin.Context)
	RunArchived(c *gin.Context)
	RunAllArchived(c *gin.Context)
	DeleteArchived(c *gin.Context)
}

type TaskHandlerImpl struct {
	logger *zap.Logger
	client *taskq.Client
}

func NewTaskHandler(logger *zap.Logger, client *taskq.Client) TaskHandler {
	return &TaskHandlerImpl{logger: logger, client: client}
}

// ListArchived godoc
// @Summary      死信任务列表
// @Description  列出所有重试耗尽已归档的任务。仅 super_admin 可访问。
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        queue  query     string  false  "队列名"  default(default)
// @Success      200    {array}   dto.ArchivedTaskItem
// @Router       /admin/tasks/archived [get]
func (h *TaskHandlerImpl) ListArchived(c *gin.Context) {
	var appCtx = ctx.FromGinCtx(c)
	queue := c.DefaultQuery("queue", "default")

	tasks, err := h.client.ListArchivedTasks(queue)
	if err != nil {
		appCtx.ToError(exception.InternalServerError.Append(err.Error()))
		return
	}

	items := make([]dto.ArchivedTaskItem, 0, len(tasks))
	for _, t := range tasks {
		items = append(items, dto.ArchivedTaskItem{
			ID:       t.ID,
			Type:     t.Type,
			Queue:    t.Queue,
			Payload:  t.Payload,
			Retried:  t.Retried,
			MaxRetry: t.MaxRetry,
			LastErr:  t.LastErr,
		})
	}
	appCtx.ToSuccess(items)
}

// RunArchived godoc
// @Summary      重试单个死信任务
// @Description  将指定已归档任务重新入队执行。仅 super_admin 可访问。
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      dto.RunTaskReq  true  "任务信息"
// @Success      200   {object}  map[string]interface{}
// @Router       /admin/tasks/archived/run [post]
func (h *TaskHandlerImpl) RunArchived(c *gin.Context) {
	var appCtx = ctx.FromGinCtx(c)
	var params dto.RunTaskReq
	if err := appCtx.ShouldBind(&params); err != nil {
		appCtx.ToError(err)
		return
	}
	if err := h.client.RunArchivedTask(params.Queue, params.TaskID); err != nil {
		appCtx.ToError(exception.InternalServerError.Append(err.Error()))
		return
	}
	appCtx.ToSuccess(nil)
}

// RunAllArchived godoc
// @Summary      重试全部死信任务
// @Description  将指定队列中所有已归档任务重新入队。仅 super_admin 可访问。
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        queue  query     string  false  "队列名"  default(default)
// @Success      200    {object}  map[string]interface{}
// @Router       /admin/tasks/archived/run-all [post]
func (h *TaskHandlerImpl) RunAllArchived(c *gin.Context) {
	var appCtx = ctx.FromGinCtx(c)
	queue := c.DefaultQuery("queue", "default")

	count, err := h.client.RunAllArchivedTasks(queue)
	if err != nil {
		appCtx.ToError(exception.InternalServerError.Append(err.Error()))
		return
	}
	appCtx.ToSuccess(map[string]int{"count": count})
}

// DeleteArchived godoc
// @Summary      删除死信任务
// @Description  永久删除指定已归档任务。仅 super_admin 可访问。
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      dto.RunTaskReq  true  "任务信息"
// @Success      200   {object}  map[string]interface{}
// @Router       /admin/tasks/archived [delete]
func (h *TaskHandlerImpl) DeleteArchived(c *gin.Context) {
	var appCtx = ctx.FromGinCtx(c)
	var params dto.RunTaskReq
	if err := appCtx.ShouldBind(&params); err != nil {
		appCtx.ToError(err)
		return
	}
	if err := h.client.DeleteArchivedTask(params.Queue, params.TaskID); err != nil {
		appCtx.ToError(exception.InternalServerError.Append(err.Error()))
		return
	}
	appCtx.ToSuccess(nil)
}
