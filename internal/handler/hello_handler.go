package handler

import (
	"go-server-starter/internal/ctx"
	"go-server-starter/internal/i18n"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type HelloHandler interface {
	Hello(c *gin.Context)
}

type HelloHandlerImpl struct {
	logger *zap.Logger
}

func NewHelloHandler(logger *zap.Logger) HelloHandler {
	return &HelloHandlerImpl{logger: logger}
}

// Hello godoc
// @Summary      健康检查
// @Description  返回 Hello World，可选 name 参数返回多语言问候
// @Tags         hello
// @Accept       json
// @Produce      json
// @Param        name  query     string  false  "名称"
// @Success      200   {object}  map[string]interface{}
// @Router       /hello [get]
func (h *HelloHandlerImpl) Hello(c *gin.Context) {
	ctx := ctx.FromGinCtx(c)
	name := c.Query("name")
	if name == "" {
		ctx.ToSuccess("Hello, World!")
		return
	}
	ctx.ToSuccess(i18n.EchoHello.Tf(ctx.GetLocale(), "name", name))
}
