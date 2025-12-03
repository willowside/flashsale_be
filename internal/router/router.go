package router

import (
	"flashsale/internal/handler"

	"github.com/gin-gonic/gin"
)

func SetupRouter(orderHandler *handler.OrderHandler) *gin.Engine {
	r := gin.Default()

	// health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Precheck endpoint
	r.POST("/flashsale/precheck", orderHandler.PreCheck)

	// TODO; add endpoints /flashsale/queue, /flashsale/stock
	return r
}
