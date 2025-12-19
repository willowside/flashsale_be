package handler

import (
	"flashsale/internal/application/dto/mq"
	"flashsale/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 1. new queue handler
// 2. feature: enqueue order message to MQ
// return order_id to client

type OrderQueueHandler struct {
	svc service.OrderQueueService
}

func NewQueueHandler(svc service.OrderQueueService) *OrderQueueHandler {
	return &OrderQueueHandler{svc: svc}
}

func (h *OrderQueueHandler) Enqueue(c *gin.Context) {
	var req mq.QueueReq
	if err := c.BindJSON(&req); err != nil || req.UserID == "" || req.ProductID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	orderID, err := h.svc.EnqueueOrder(c, req.UserID, req.ProductID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "queued",
		"order_id": orderID,
	})
}
