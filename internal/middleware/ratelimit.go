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

// RateLimitByUser limits by authenticated user ID (uniCode from JWT). Falls back to IP.
func (r *RateLimit) RateLimitByUser(rate int, zones ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		appCtx := ctx.FromGinCtx(c)
		key := userKey(c, zones)
		if !r.check(appCtx, key, rate, zones) {
			return
		}
		c.Next()
	}
}

// RateLimit limits by IP address.
func (r *RateLimit) RateLimit(rate int, zones ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		appCtx := ctx.FromGinCtx(c)
		key := ipKey(c, zones)
		if !r.check(appCtx, key, rate, zones) {
			return
		}
		c.Next()
	}
}

func ipKey(c *gin.Context, zones []string) string {
	return fmt.Sprintf("ip:%s", constant.RedisKeyOfRateLimit(zoneStr(zones), c.ClientIP()))
}

func userKey(c *gin.Context, zones []string) string {
	if uniCode := c.GetString(constant.CTX_KEY_OF_USER_UNI_CODE); uniCode != "" {
		return fmt.Sprintf("user:%s", constant.RedisKeyOfRateLimit(zoneStr(zones), uniCode))
	}
	return ipKey(c, zones) // fallback to IP before JWT middleware runs
}

func zoneStr(zones []string) string {
	if len(zones) == 0 {
		return "GLOBAL"
	}
	return strings.ToUpper(strings.Join(zones, ":"))
}

func (r *RateLimit) check(appCtx *ctx.Context, key string, rate int, zones []string) bool {
	limiter := redis_rate.NewLimiter(r.redis)
	res, err := limiter.Allow(appCtx.Ctx, key, redis_rate.PerMinute(rate))
	if err != nil {
		r.logger.Warn("redis rate limit failed, using local fallback", zap.Error(err))
		if !r.localAllow(key, rate) {
			appCtx.ToError(exception.TooManyRequests)
			return false
		}
		return true
	}

	headerKey := fmt.Sprintf("X-RateLimit-%s-Remaining", zoneStr(zones))
	appCtx.Gtx.Header(headerKey, strconv.FormatInt(int64(res.Remaining), 10))

	if res.Allowed == 0 {
		appCtx.Gtx.Header("Retry-After", strconv.FormatFloat(res.RetryAfter.Seconds(), 'f', 0, 64))
		appCtx.ToError(exception.TooManyRequests)
		return false
	}
	return true
}

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
