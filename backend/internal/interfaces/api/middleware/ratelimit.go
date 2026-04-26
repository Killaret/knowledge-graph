package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter implements token bucket algorithm for rate limiting
type RateLimiter struct {
	requests map[string][]time.Time
	mu       sync.RWMutex
	limit    int
	window   time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// Allow checks if request from client is allowed
func (rl *RateLimiter) Allow(clientID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	requests := rl.requests[clientID]

	// Remove old requests outside the window
	valid := make([]time.Time, 0)
	for _, req := range requests {
		if now.Sub(req) < rl.window {
			valid = append(valid, req)
		}
	}

	// Check if limit exceeded
	if len(valid) >= rl.limit {
		rl.requests[clientID] = valid
		return false
	}

	// Add current request
	valid = append(valid, now)
	rl.requests[clientID] = valid
	return true
}

// RateLimitMiddleware creates gin middleware for rate limiting
func RateLimitMiddleware(limit int, window time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(limit, window)

	return func(c *gin.Context) {
		clientID := c.ClientIP()
		if !limiter.Allow(clientID) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Please try again later.",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// RateLimitByEndpoint creates rate limiter per endpoint
func RateLimitByEndpoint(limits map[string]int, defaultLimit int, window time.Duration) gin.HandlerFunc {
	limiters := make(map[string]*RateLimiter)
	for endpoint, limit := range limits {
		limiters[endpoint] = NewRateLimiter(limit, window)
	}
	defaultLimiter := NewRateLimiter(defaultLimit, window)

	return func(c *gin.Context) {
		clientID := c.ClientIP()
		endpoint := c.FullPath()

		var limiter *RateLimiter
		if l, ok := limiters[endpoint]; ok {
			limiter = l
		} else {
			limiter = defaultLimiter
		}

		if !limiter.Allow(clientID + ":" + endpoint) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded for this endpoint",
				"retry_after": window.Seconds(),
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
