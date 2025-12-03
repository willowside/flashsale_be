package handler

import (
	"flashsale/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	svc *service.OrderService
}

func NewOrderHandler(svc *service.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

// POST /flashsale/precheck?user_id=<USER_ID>&product_id=<PRODUCT_ID>
func (h *OrderHandler) PreCheck(c *gin.Context) {
	ctx := c.Request.Context()

	// query string or JSON body
	userID := c.Query("user_id")
	productID := c.Query("product_id")

	// try parse the body if empty query
	if userID == "" || productID == "" {
		var body struct {
			UserID    string `json:"user_id"`
			ProductID string `json:"product_id"`
		}
		if err := c.BindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing user_id or product_id"})
			return
		}
		userID = body.UserID
		productID = body.ProductID
	}

	if userID == "" || productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id & product_id are required"})
		return
	}

	result, err := h.svc.PreCheckAndQueue(ctx, userID, productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  result.Status,
			"message": result.Message,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  result.Status,
		"message": result.Message,
	})

}
