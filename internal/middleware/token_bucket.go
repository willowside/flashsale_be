package middleware

import (
	"flashsale/internal/cache"
	"net/http"

	"github.com/gin-gonic/gin"
)

var TB *cache.TokenBucketLimiter

// Init token bucket limiter
func InitTokenBucketLimiter(tb *cache.TokenBucketLimiter) {
	TB = tb
}

// User based token bucket
func UserTokenBucket(maxTokens, refillPerSec float64) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID := ctx.GetHeader("X-User-ID")
		if userID == "" {
			// no user -> fallback to IP
			ctx.Next()
			return
		}

		// redis keyPrefix
		keyPrefix := "tb:user:" + userID
		allowed, _, err := TB.Allow(keyPrefix, maxTokens, refillPerSec)
		if err != nil {
			// fail open
			ctx.Next()
			return
		}

		if !allowed {
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "user rate limit blocked"})
			return
		}
		ctx.Next()
	}
}

// IP token bucket
func IPTokenBucket(maxTokens, refillPerSec float64) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ip := ctx.ClientIP()
		kexPrefix := "tb:ip:" + ip
		allowed, _, err := TB.Allow(kexPrefix, maxTokens, refillPerSec)
		if err != nil {
			ctx.Next()
			return
		}
		if !allowed {
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "IP rate limit blocked"})
			return
		}
		ctx.Next()
	}
}
