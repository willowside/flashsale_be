package service

import (
	"context"
	"flashsale/internal/dto"
	"flashsale/internal/repository/repositoryiface"
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

func (c *OrderCompensator) Compensate(ctx context.Context, msg dto.DLQMessage) {
	// 1. mark order failed
	if err := c.orderRepo.MarkOrderFailed(ctx, msg.OrderNo, msg.Reason); err != nil {
		log.Printf("[DLQ] mark failed order=%s, err=%v", msg.OrderNo, err)
	}
	// 2. restore stock
	if err := c.redisStockRepo.RestoreStock(ctx, strconv.FormatInt(msg.Payload.ProductID, 10), 1); err != nil {
		log.Printf("[DLQ] restore stock failed order =%s, product=%d, err=%v", msg.OrderNo, msg.Payload.ProductID, err)
	}

}
