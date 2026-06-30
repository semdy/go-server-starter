package middleware

import (
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// sensitiveHeaders lists headers that should be redacted from logs.
var sensitiveHeaders = map[string]bool{
	"Authorization": true,
	"Cookie":        true,
	"Set-Cookie":    true,
	"X-Api-Key":     true,
}

func ZapLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		// Skip allocation if Info level is not enabled (e.g. production at Warn)
		if ce := logger.Check(zap.InfoLevel, "request"); ce != nil {
			timeSpend := time.Since(start)
			fields := []zap.Field{
				zap.Int("status", c.Writer.Status()),
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.String("query", c.Request.URL.RawQuery),
				zap.String("ip", c.ClientIP()),
				zap.String("user-agent", c.Request.UserAgent()),
				zap.Float64("time_spend_ms", float64(timeSpend.Microseconds())/1000.0),
			}

			// Only build header map when actually logging
			if len(c.Request.Header) > 0 {
				headersMap := make(map[string]string, len(c.Request.Header))
				for k, v := range c.Request.Header {
					if sensitiveHeaders[k] {
						headersMap[k] = "[REDACTED]"
					} else {
						headersMap[k] = strings.Join(v, ",")
					}
				}
				fields = append(fields, zap.Any("headers", headersMap))
			}

			ce.Write(fields...)
		}
	}
}

// ZapRecovery 恢复中间件
func ZapRecovery(logger *zap.Logger, stack bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				fields := []zap.Field{
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.String("ip", c.ClientIP()),
				}

				// 根据 stack 参数决定是否记录堆栈信息
				if stack {
					fields = append(fields, zap.String("stack", string(debug.Stack())))
				}

				logger.Error("panic recovered", fields...)
				c.AbortWithStatus(500)
			}
		}()
		c.Next()
	}
}
