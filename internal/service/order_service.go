package service

import (
	"context"
	"flashsale/internal/cache"
	"flashsale/internal/repository/repositoryiface"
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
	repo      repositoryiface.OrderRepository
	publisher serviceiface.OrderPublisher
}

func NewOrderService(pub serviceiface.OrderPublisher, lua *cache.LuaScripts, repo repositoryiface.OrderRepository) *OrderService {
	return &OrderService{
		repo:      repo,
		publisher: pub,
		lua:       lua,
	}
}

func (s *OrderService) PreCheckAndQueue(ctx context.Context, userID, productID string) (*PrecheckResult, error) {
	// 1. redis precheck
	res, err := cache.FlashSalePreCheck(productID, userID)
	if err != nil {
		return &PrecheckResult{Status: "error", Message: err.Error()}, err
	}

	if !res.Success {
		return &PrecheckResult{
			Status:  res.Reason,
			Message: "precheck failed",
		}, nil

		// switch res.Reason {
		// case "OUT_OF_STOCK", "STOCK_NOT_FOUND":
		// 	return &PrecheckResult{Status: "no_stock", Message: "product out of stock"}, nil
		// case "USER_ALREADY_PURCHASED", "USER_ALREADY_BOUGHT":
		// 	return &PrecheckResult{Status: "already_purchased", Message: "user already purchased"}, nil
		// default:
		// 	return &PrecheckResult{Status: "error", Message: "precheck failed: " + res.Reason}, nil
		// }
	}
	// 2. gen OrderID
	orderID := uuid.New().String()

	// 3. create PENDING order in DB
	err = s.repo.CreatePendingOrder(ctx, orderID, userID, productID)
	if err != nil {
		return nil, fmt.Errorf("create pending order failed: %w", err)
	}

	// 4. publish MQ
	if err := s.publisher.PublishOrder(ctx, orderID, userID, productID); err != nil {
		return nil, fmt.Errorf("[order service] publish order failed: %w", err)
	}

	return &PrecheckResult{
		Status:  "queued",
		OrderID: orderID,
		Message: "precheck success, order queued",
	}, nil
}
