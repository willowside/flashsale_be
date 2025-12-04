package router

import (
	"flashsale/internal/handler"
	"flashsale/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter(orderHandler *handler.OrderHandler) *gin.Engine {
	r := gin.Default()

	// global limiter
	r.Use(middleware.GlobalRateLimit(1)) // cap 2000 req/s

	// IP limiter globally
	r.Use(middleware.IPRateLimit())

	// health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Precheck endpoint
	r.POST("/flashsale/precheck", middleware.UserRateLimit(), orderHandler.PreCheck)

	// TODO; add endpoints /flashsale/queue, /flashsale/stock
	return r
}
