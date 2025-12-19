package middleware

import (
	"flashsale/internal/cache"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// 1. init global hybrid rate limiter
// 2. ip limiter + user limiter + global limiter

// global hybrid rate limiter instance
var HybridLimiter *cache.HybridLimiter

func InitHybridLimiter(h *cache.HybridLimiter) {
	HybridLimiter = h
}

// wrapper params: capacity(burst), refillRate(tokens/sec), slidingLimit()

func UserHybridLimiter(capacity, refillPerSec float64, slidingLimit int64, windowSec int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. get user id from header X-User-ID -> else fallback to query
		userID := c.GetHeader("X-User-ID")
		if userID == "" {
			userID = c.Query("user_id")
		}
		if userID == "" {
			// no user id -> then fallback to ip limiter/ Allow
			c.Next()
			return
		}
		// 2. call HybridLimiter.Allow
		allowed, reason, tokensLeft, swCount, err := HybridLimiter.Allow(c.Request.Context(), "user:"+userID, capacity, refillPerSec, slidingLimit, windowSec)
		if err != nil {
			// fail open
			c.Next()
			return
		}

		// debug headers
		c.Header("X-Rate-Allowed", strconv.FormatBool(allowed))
		c.Header("X-Rate-Reason", reason)
		c.Header("X-Rate-Tokens", strconv.FormatFloat(tokensLeft, 'f', 2, 64))
		c.Header("X-Rate-Sliding-Count", strconv.FormatInt(swCount, 10))

		if !allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":  "rate limit exceeded",
				"reason": reason,
			})
			return
		}
		c.Next()
	}

}

func IPHybridLimiter(capacity, refillPerSec float64, slidingLimit int64, windowSec int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		allowed, reason, tokens, sw, err := HybridLimiter.Allow(c.Request.Context(), "ip:"+ip, capacity, refillPerSec, slidingLimit, windowSec)
		if err != nil {
			c.Next()
			return
		}
		c.Header("X-Rate-Allowed", strconv.FormatBool(allowed))
		c.Header("X-Rate-Reason", reason)
		c.Header("X-Rate-Tokens", strconv.FormatFloat(tokens, 'f', 2, 64))
		c.Header("X-Rate-Sliding", strconv.FormatInt(sw, 10))
		if !allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded", "reason": reason})
			return
		}
		c.Next()
	}
}

// Global limiter: use a single key "global"
func GlobalHybridLimiter(capacity, refillPerSec float64, slidingLimit int64, windowSec int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		allowed, reason, _, _, err := HybridLimiter.Allow(c.Request.Context(), "global", capacity, refillPerSec, slidingLimit, windowSec)
		if err != nil {
			c.Next()
			return
		}
		c.Header("X-Rate-Allowed", strconv.FormatBool(allowed))
		c.Header("X-Rate-Reason", reason)
		if !allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "global rate limit exceeded", "reason": reason})
			return
		}
		c.Next()
	}
}
