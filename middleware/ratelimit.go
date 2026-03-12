package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type clientLimit struct {
	count     int
	resetTime time.Time
}

var (
	clients = make(map[string]*clientLimit)
	mu      sync.Mutex
)

const (
	limit        = 180
	windowPeriod = 1 * time.Minute
)

func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		now := time.Now()

		mu.Lock()
		client, exists := clients[clientIP]

		if !exists || now.After(client.resetTime) {
			client = &clientLimit{
				count:     0,
				resetTime: now.Add(windowPeriod),
			}
			clients[clientIP] = client
		}

		client.count++
		remaining := limit - client.count
		if remaining < 0 {
			remaining = 0
		}

		// Set rate limit headers
		c.Header("x-ratelimit-limit", fmt.Sprintf("%d", limit))
		c.Header("x-ratelimit-remaining", fmt.Sprintf("%d", remaining))
		c.Header("x-ratelimit-reset", fmt.Sprintf("%d", client.resetTime.Unix()))
		c.Header("x-ratelimit-global", "true")

		if client.count > limit {
			mu.Unlock()
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Please try again later.",
			})
			c.Abort()
			return
		}

		mu.Unlock()
		c.Next()
	}
}
