package router

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type clientRateState struct {
	tokens float64
	last   time.Time
}

// RateLimitByClientIP is a bounded in-process token bucket. The reverse proxy
// remains the outer DDoS layer; this protects expensive application routes and
// store verification from accidental bursts and simple abuse.
func RateLimitByClientIP(perMinute, burst int) gin.HandlerFunc {
	var mu sync.Mutex
	clients := make(map[string]clientRateState)
	ratePerSecond := float64(perMinute) / 60
	lastCleanup := time.Now()

	return func(c *gin.Context) {
		now := time.Now()
		key := c.ClientIP()
		mu.Lock()
		state, exists := clients[key]
		if !exists {
			state = clientRateState{tokens: float64(burst), last: now}
		}
		state.tokens += now.Sub(state.last).Seconds() * ratePerSecond
		if state.tokens > float64(burst) {
			state.tokens = float64(burst)
		}
		state.last = now
		allowed := state.tokens >= 1
		if allowed {
			state.tokens--
		}
		clients[key] = state
		if now.Sub(lastCleanup) >= 5*time.Minute {
			for client, item := range clients {
				if now.Sub(item.last) > 10*time.Minute {
					delete(clients, client)
				}
			}
			lastCleanup = now
		}
		mu.Unlock()

		if !allowed {
			c.Header("Retry-After", "5")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": true, "message": "Too many requests; please try again shortly",
			})
			return
		}
		c.Next()
	}
}
