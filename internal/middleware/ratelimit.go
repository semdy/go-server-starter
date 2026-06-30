package middleware

import (
	"fmt"
	"go-server-starter/internal/constant"
	"go-server-starter/internal/ctx"
	"go-server-starter/internal/exception"
	"go-server-starter/pkg/redis"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis_rate/v10"
	"go.uber.org/zap"
)

type RateLimit struct {
	redis  *redis.Client
	logger *zap.Logger

	// local fallback when Redis is down
	mu          sync.Mutex
	localCounts map[string]*localWindow
}

type localWindow struct {
	count   int
	started time.Time
}

func NewRateLimit(redis *redis.Client, logger *zap.Logger) *RateLimit {
	rl := &RateLimit{
		redis:       redis,
		logger:      logger,
		localCounts: make(map[string]*localWindow),
	}
	// Periodically purge stale local counters
	go rl.purgeLoop()
	return rl
}

func (r *RateLimit) purgeLoop() {
	for range time.Tick(2 * time.Minute) {
		r.purgeStale()
	}
}

func (r *RateLimit) purgeStale() {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	for k, w := range r.localCounts {
		if now.Sub(w.started) > 2*time.Minute {
			delete(r.localCounts, k)
		}
	}
}

// RateLimit 限流中间件（每分钟限流 rate 次，Redis GCRA + 本地降级）
func (r *RateLimit) RateLimit(rate int, zones ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		appCtx := ctx.FromGinCtx(c)
		ip := c.ClientIP()

		if len(zones) == 0 {
			zones = []string{"GLOBAL"}
		}
		zoneStr := strings.ToUpper(strings.Join(zones, ":"))
		key := constant.RedisKeyOfRateLimit(zoneStr, ip)

		limiter := redis_rate.NewLimiter(r.redis)
		res, err := limiter.Allow(appCtx.Ctx, key, redis_rate.PerMinute(rate))
		if err != nil {
			// Redis unreachable — fall back to local sliding window
			r.logger.Warn("redis rate limit failed, using local fallback", zap.Error(err))
			if !r.localAllow(key, rate) {
				appCtx.ToError(exception.TooManyRequests)
				return
			}
			c.Next()
			return
		}

		headerKey := fmt.Sprintf("X-RateLimit-%s-Remaining", zoneStr)
		c.Header(headerKey, strconv.FormatInt(int64(res.Remaining), 10))

		if res.Allowed == 0 {
			c.Header("Retry-After", strconv.FormatFloat(res.RetryAfter.Seconds(), 'f', 0, 64))
			appCtx.ToError(exception.TooManyRequests)
			return
		}
		c.Next()
	}
}

// localAllow implements a simple per-minute sliding window in memory.
func (r *RateLimit) localAllow(key string, rate int) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	w, exists := r.localCounts[key]
	if !exists || now.Sub(w.started) > time.Minute {
		r.localCounts[key] = &localWindow{count: 1, started: now}
		return true
	}
	if w.count >= rate {
		return false
	}
	w.count++
	return true
}
