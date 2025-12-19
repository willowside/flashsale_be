package handler

import (
	"flashsale/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type StockHandler struct {
	svc service.StockService
}

func NewStockHandler(svc service.StockService) *StockHandler {
	return &StockHandler{svc: svc}
}

func (h *StockHandler) GetStock(c *gin.Context) {
	productId := c.Param("product_id")
	pid, err := strconv.ParseInt(productId, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product_id"})
		return
	}

	stock, err := h.svc.GetStock(c.Request.Context(), pid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"product_id": pid, "stock": stock})
}
