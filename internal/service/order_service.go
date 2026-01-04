package service

import (
	"context"
	"flashsale/internal/cache"
	"flashsale/internal/repository/repositoryiface"
	"flashsale/internal/service/serviceiface"
	"fmt"
	"time"

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
	lua           *cache.LuaScripts
	repo          repositoryiface.OrderRepository
	publisher     serviceiface.OrderPublisher
	flashSaleRepo repositoryiface.FlashSaleRepository
}

func NewOrderService(pub serviceiface.OrderPublisher, lua *cache.LuaScripts, repo repositoryiface.OrderRepository, flashSaleRepo repositoryiface.FlashSaleRepository) *OrderService {
	return &OrderService{
		repo:          repo,
		publisher:     pub,
		lua:           lua,
		flashSaleRepo: flashSaleRepo,
	}
}

func (s *OrderService) PreCheckAndQueue(ctx context.Context, userID, productID string) (*PrecheckResult, error) {
	// 0 Flash Sale Window Gate
	fs, err := s.flashSaleRepo.GetActiveFlashSale(ctx)
	if err != nil || fs == nil {
		return &PrecheckResult{
			Status:  "not_started",
			Message: "flash sale not active",
		}, nil
	}

	if !fs.IsActive(time.Now()) {
		return &PrecheckResult{
			Status:  "not_started",
			Message: "flash sale not in window",
		}, nil
	}

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
	}
	// get promotional price from flash_sale_products table
	fsp, err := s.flashSaleRepo.GetFlashSaleProduct(ctx, fs.ID, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch flash sale product details: %w", err)
	}

	// 2. gen OrderID
	orderID := uuid.New().String()

	// 3. create PENDING order in DB
	err = s.repo.CreatePendingOrder(ctx, orderID, userID, productID, fs.ID, fsp.SalePrice)
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
