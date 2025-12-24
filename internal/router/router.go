package router

import (
	"flashsale/internal/cache"
	"flashsale/internal/handler"
	"flashsale/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter(orderHandler *handler.OrderHandler, stockHandler *handler.StockHandler, resultHandler *handler.OrderResultHandler) *gin.Engine {
	r := gin.Default()

	// // load token bucket lua
	// tb, err := cache.LoadTokenBucketLua("./scripts/rate_limit_token_bucket.lua")
	// if err != nil {
	// 	panic(err)
	// }

	// middleware.InitTokenBucketLimiter(tb)

	// load hybrid lua
	hl, err := cache.LoadHybridLimiter(cache.Rdb, "./scripts/rate_limit_hybrid.lua")
	if err != nil {
		panic(err)
	}
	middleware.InitHybridLimiter(hl)

	// health check
	// r.GET("/health", func(c *gin.Context) {
	// 	c.JSON(200, gin.H{"status": "ok"})
	// })

	// Precheck endpoint
	// r.POST("/flashsale/precheck", middleware.UserRateLimit(), orderHandler.PreCheck)

	flash := r.Group("/flashsale")
	{
		flash.POST("/precheck",
			middleware.UserHybridLimiter(20, 10, 50, 5),
			// middleware.UserTokenBucket(5, 3), // burst: 5, refill: 3 tokens/sec
			orderHandler.PreCheck)
		flash.POST("/stock/:product_id",
			middleware.UserHybridLimiter(20, 10, 50, 5),
			stockHandler.GetStock)
		flash.GET("/result/:order_id",
			middleware.UserHybridLimiter(5, 2, 10, 1),
			resultHandler.GetResult)

	}

	// TODO; add endpoints /flashsale/queue, /flashsale/stock
	return r
}
