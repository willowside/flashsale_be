package service

import (
	"context"
	"flashsale/internal/cache"
	"flashsale/internal/service/serviceiface"
	"fmt"

	"github.com/google/uuid"
)

/*

	DTO

*/

// PrecheckResult encapsulates results to handler
type PrecheckResult struct {
	Status  string // queued/ no_stock/ already_purchased
	OrderID string
	Message string
}

/*

	SERVICE

*/

// OrderService: Precheck logic + put request into Queue
type OrderService struct {
	lua       *cache.LuaScripts
	publisher serviceiface.OrderPublisher
	ttlSec    int // ttl for per user lock (DI from config or main)
}

func (s *OrderService) PreCheckAndQueue(ctx context.Context, userID, productID string) (*PrecheckResult, error) {
	res, err := cache.FlashSalePreCheck(productID, userID)
	if err != nil {
		return &PrecheckResult{Status: "error", Message: fmt.Sprintf("precheck error: %v", err)}, err
	}

	if !res.Success {
		switch res.Reason {
		case "OUT_OF_STOCK", "STOCK_NOT_FOUND":
			return &PrecheckResult{Status: "no_stock", Message: "product out of stock"}, nil
		case "USER_ALREADY_PURCHASED", "USER_ALREADY_BOUGHT":
			return &PrecheckResult{Status: "already_purchased", Message: "user already purchased"}, nil
		default:
			return &PrecheckResult{Status: "error", Message: "precheck failed: " + res.Reason}, nil
		}
	}

	if s.publisher == nil {
		return &PrecheckResult{Status: "queued", Message: "precheck success, dev-mode no publisher"}, nil
	}

	orderID := uuid.New().String()

	if err := s.publisher.PublishOrder(ctx, orderID, userID, productID); err != nil {
		return &PrecheckResult{Status: "error", Message: fmt.Sprintf("failed to enqueue order: %v", err)}, err
	}

	return &PrecheckResult{
		Status:  "queued",
		OrderID: orderID,
		Message: "precheck success, order queued",
	}, nil
}

func NewOrderService(pub serviceiface.OrderPublisher, lua *cache.LuaScripts, ttlSeconds int) *OrderService {
	return &OrderService{
		publisher: pub,
		lua:       lua,
		ttlSec:    ttlSeconds,
	}
}
