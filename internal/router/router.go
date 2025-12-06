package router

import (
	"flashsale/internal/cache"
	"flashsale/internal/handler"
	"flashsale/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter(orderHandler *handler.OrderHandler) *gin.Engine {
	r := gin.Default()

	// load token bucket lua
	tb, err := cache.LoadTokenBucketLua("internal/cache/lua/ratelimit_token_bucket.lua")
	if err != nil {
		panic(err)
	}

	middleware.InitTokenBucketLimiter(tb)

	// health check
	// r.GET("/health", func(c *gin.Context) {
	// 	c.JSON(200, gin.H{"status": "ok"})
	// })

	// Precheck endpoint
	// r.POST("/flashsale/precheck", middleware.UserRateLimit(), orderHandler.PreCheck)

	flash := r.Group("/flashsale")
	{
		flash.POST("/precheck",
			middleware.UserTokenBucket(5, 3), // burst: 5, refill: 3 tokens/sec
			orderHandler.PreCheck)
	}

	// TODO; add endpoints /flashsale/queue, /flashsale/stock
	return r
}
