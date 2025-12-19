package handler

import (
	"flashsale/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type OrderResultHandler struct {
	svc *service.OrderResultService
}

func NewOrderResultHandler(svc *service.OrderResultService) *OrderResultHandler {
	return &OrderResultHandler{svc: svc}
}

// GET /flashsale/result/:order_id
func (h *OrderResultHandler) GetResult(c *gin.Context) {
	orderID := c.Param("order_id")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order_id required"})
		return
	}

	result, err := h.svc.GetResult(c.Request.Context(), orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
