package api

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateLimiter struct {
	visits map[string][]time.Time
	mu     sync.Mutex
	limit  int
	window time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{
		visits: make(map[string][]time.Time),
		limit:  limit,
		window: window,
	}
	go rl.cleanup()
	return rl
}

func (rl *rateLimiter) cleanup() {
	for {
		time.Sleep(rl.window)
		rl.mu.Lock()
		now := time.Now()
		for ip, times := range rl.visits {
			filtered := times[:0]
			for _, t := range times {
				if now.Sub(t) < rl.window {
					filtered = append(filtered, t)
				}
			}
			if len(filtered) == 0 {
				delete(rl.visits, ip)
			} else {
				rl.visits[ip] = filtered
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *rateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		rl.mu.Lock()
		times := rl.visits[ip]
		now := time.Now()
		// Remove old timestamps
		var filtered []time.Time
		for _, t := range times {
			if now.Sub(t) < rl.window {
				filtered = append(filtered, t)
			}
		}
		if len(filtered) >= rl.limit {
			rl.mu.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded",
				"code":    "ERR_RATE_LIMIT",
				"details": "Too many requests from this IP",
				"docs":    "https://docs.kubechat.dev/errors/rate-limit",
			})
			return
		}
		filtered = append(filtered, now)
		rl.visits[ip] = filtered
		rl.mu.Unlock()
		c.Next()
	}
}
