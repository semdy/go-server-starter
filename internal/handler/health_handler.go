package handler

import (
	"context"
	"time"

	"go-server-starter/internal/ctx"
	"go-server-starter/pkg/database"
	"go-server-starter/pkg/redis"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type HealthHandler interface {
	Healthz(c *gin.Context)
}

type HealthHandlerImpl struct {
	logger *zap.Logger
	db     *database.DB
	redis  *redis.Client
}

func NewHealthHandler(logger *zap.Logger, db *database.DB, redis *redis.Client) HealthHandler {
	return &HealthHandlerImpl{logger: logger, db: db, redis: redis}
}

// Healthz godoc
// @Summary      健康检查
// @Description  返回服务及依赖（DB、Redis）的健康状态
// @Tags         health
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      503  {object}  map[string]interface{}
// @Router       /healthz [get]
func (h *HealthHandlerImpl) Healthz(c *gin.Context) {
	var appCtx = ctx.FromGinCtx(c)
	status := "ok"
	details := map[string]string{}

	ctx, cancel := context.WithTimeout(appCtx.Ctx, 3*time.Second)
	defer cancel()

	if h.db != nil {
		if err := h.db.Ping(ctx); err != nil {
			status = "degraded"
			details["database"] = "unreachable: " + err.Error()
		} else {
			details["database"] = "ok"
		}
	} else {
		details["database"] = "not configured"
	}

	if h.redis != nil {
		if err := h.redis.Ping(ctx).Err(); err != nil {
			status = "degraded"
			details["redis"] = "unreachable: " + err.Error()
		} else {
			details["redis"] = "ok"
		}
	} else {
		details["redis"] = "not configured"
	}

	httpStatus := 200
	if status != "ok" {
		httpStatus = 503
	}
	appCtx.Gtx.JSON(httpStatus, gin.H{"status": status, "details": details})
}
