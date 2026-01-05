package service

import (
	"context"
	"flashsale/internal/dto"
	"flashsale/internal/repository/repositoryiface"
	"fmt"
	"log"
	"strconv"
)

type OrderCompensator struct {
	orderRepo      repositoryiface.OrderRepository
	redisStockRepo repositoryiface.RedisStockRepository
}

func NewOrderCompensator(
	orderRepo repositoryiface.OrderRepository,
	redisStockRepo repositoryiface.RedisStockRepository,
) *OrderCompensator {
	return &OrderCompensator{
		orderRepo:      orderRepo,
		redisStockRepo: redisStockRepo,
	}
}

func (c *OrderCompensator) Compensate(ctx context.Context, msg dto.DLQMessage) error {
	log.Printf("[Compensator] start compensate: OrderNo=%s, Reason=%s", msg.OrderNo, msg.Reason)
	// 1. Idempotency: check DB order status
	status, err := c.orderRepo.GetOrderStatus(ctx, msg.OrderNo)
	if err != nil {
		return err
	}

	// if already SUCCESS/FAILED, no compensation
	if status == "success" || status == "failed" {
		log.Printf("[Compensator] order %s already processed, skip compensation", msg.OrderNo)
		return nil
	}

	// 2. mark order failed
	if err := c.orderRepo.MarkOrderFailed(ctx, msg.OrderNo, msg.Reason); err != nil {
		return fmt.Errorf("failed to mark order as failed=%s, err=%v", msg.OrderNo, err)
	}
	// 3. restore stock
	productID := msg.Payload.ProductID
	log.Printf("[Compensator] run redis stock compensate: ProductID=%d", productID)
	if err := c.redisStockRepo.RestoreStock(ctx, strconv.FormatInt(msg.Payload.ProductID, 10), 1); err != nil {
		log.Printf("[Compensator ERROR] restore stock failed order =%s, product=%d, err=%v", msg.OrderNo, msg.Payload.ProductID, err)
		return err
	}

	log.Printf("[Compensator] compensation success: OrderNo=%s", msg.OrderNo)
	return nil

}
