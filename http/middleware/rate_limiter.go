package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter 限流器结构体
type RateLimiter struct {
	mu       sync.RWMutex
	requests map[string]*RequestRecord
	limit    int           // 每秒请求数限制
	window   time.Duration // 时间窗口
}

// RequestRecord 请求记录
type RequestRecord struct {
	count    int
	lastTime time.Time
}

// NewRateLimiter 创建新的限流器
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string]*RequestRecord),
		limit:    limit,
		window:   window,
	}

	// 定期清理过期记录的 goroutine
	go rl.cleanup()

	return rl
}

// cleanup 定期清理过期的请求记录
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(1 * time.Minute) // 每分钟清理一次
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, record := range rl.requests {
			if now.Sub(record.lastTime) > rl.window {
				delete(rl.requests, key)
			}
		}
		rl.mu.Unlock()
	}
}

// Limit 限流中间件
func (rl *RateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 使用客户端IP作为限流标识
		clientIP := c.ClientIP()

		rl.mu.Lock()
		record, exists := rl.requests[clientIP]
		if !exists {
			record = &RequestRecord{
				count:    1,
				lastTime: time.Now(),
			}
			rl.requests[clientIP] = record
		} else {
			now := time.Now()
			// 如果时间窗口已过，重置计数器
			if now.Sub(record.lastTime) >= rl.window {
				record.count = 1
				record.lastTime = now
			} else {
				record.count++
				record.lastTime = now
			}
		}
		rl.mu.Unlock()

		// 检查是否超过限制
		if record.count > rl.limit {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "请求过于频繁，请稍后再试",
				"code":  429,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GlobalRateLimiter 全局限流器实例，限制为每秒10个请求
var GlobalRateLimiter = NewRateLimiter(10, 1*time.Second)
