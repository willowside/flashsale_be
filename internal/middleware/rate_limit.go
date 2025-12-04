package middleware

import (
	"context"
	"flashsale/internal/cache"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// fixed window: per-user-per-second rate limiter, use Redis INCR+EXPIRE

// Basic settings
var (
	// requests per user per second
	UserLimitPerSecond = int64(5)
	// requests per IP per second
	IPLimitPerSecond  = int64(20)
	AllowOnRedisError = true
)

// incrWithExpire does INCR & set EXPIRE when first created
func incrWithExpire(ctx context.Context, key string, expireSec int) (int64, error) {
	// redis do INCR
	n, err := cache.Rdb.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	// if first time, set expire
	if n == 1 {
		cache.Rdb.Expire(ctx, key, time.Duration(expireSec)*time.Second)
	}
	return n, nil
}

// UserRateLimit middleware uses X-User-ID header
func UserRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		// get user ID from header
		userID := c.GetHeader("X-User-ID")
		if userID == "" {
			userID = c.Query("user_id")
		}
		if userID == "" {
			// no user provided, treat as anonymous user -> use IP rate limit
			c.Next()
			return
		}

		// build redis key
		key := fmt.Sprintf("ratelimit:user:%s:%d", userID, time.Now().Unix())
		n, err := incrWithExpire(ctx, key, 1)
		if err != nil {
			if AllowOnRedisError {
				c.Next()
				return
			}
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit check failed"})
			return
		}

		if n > UserLimitPerSecond {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "user rate limit exceeded"})
			return
		}
		c.Next()
	}
}

// IPRateLimit middleware uses request IP address
func IPRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		// get client IP
		ip := c.ClientIP()
		// build redis key
		key := fmt.Sprintf("ratelimit:ip:%s:%d", ip, time.Now().Unix())
		n, err := incrWithExpire(ctx, key, 1)
		if err != nil {
			if AllowOnRedisError {
				c.Next()
				return
			}
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "ip rate limit check failed"})
			return
		}
		if n > IPLimitPerSecond {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "ip rate limit exceeded"})
			return
		}
		c.Next()
	}
}

// simple global limiter
func GlobalRateLimit(maxPerSecond int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		// build redis key
		key := fmt.Sprintf("ratelimit:global:%d", time.Now().Unix())
		n, err := incrWithExpire(ctx, key, 1)
		if err != nil {
			if AllowOnRedisError {
				c.Next()
				return
			}
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "global rate limit check failed"})
			return
		}
		if n > maxPerSecond {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "global rate limit exceeded"})
			return
		}
		c.Next()
	}
}
