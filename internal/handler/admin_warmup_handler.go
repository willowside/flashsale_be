// internal/handler/admin_warmup_handler.go
package handler

import (
	"net/http"
	"strconv"

	"flashsale/internal/service"

	"github.com/gin-gonic/gin"
)

type WarmUpHandler struct {
	svc *service.FlashSaleWarmUpService
}

func NewWarmUpHandler(svc *service.FlashSaleWarmUpService) *WarmUpHandler {
	return &WarmUpHandler{svc: svc}
}

func (h *WarmUpHandler) WarmUp(c *gin.Context) {
	// 1. get param ID: /admin/flashsales/:id/warmup
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid flash sale id"})
		return
	}

	// 2. call servicd
	if err := h.svc.WarmUpByID(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "warm up successful for flash sale " + idStr,
	})
}
