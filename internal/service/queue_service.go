package service

import (
	"context"
	"flashsale/internal/service/serviceiface"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type OrderQueueService interface {
	EnqueueOrder(ctx context.Context, userID, productID string) (string, error)
}

type orderQueueService struct {
	publisher serviceiface.OrderPublisher
	rdb       *redis.Client // temp order status
}

func NewQueueService(rdb *redis.Client, publisher serviceiface.OrderPublisher) OrderQueueService {
	return &orderQueueService{rdb: rdb, publisher: publisher}
}

func (qs *orderQueueService) EnqueueOrder(ctx context.Context, userID, productID string) (string, error) {
	orderID := uuid.New().String()

	// 1. push into RabbitMQ (worker then consume)
	if err := qs.publisher.PublishOrder(ctx, orderID, userID, productID); err != nil {
		return "", fmt.Errorf("failed to publish queue: %w", err)
	}

	// 2. init order status to redis (pending)
	key := fmt.Sprintf("flashsale:order:%s", orderID)
	if err := qs.rdb.HSet(ctx, key, map[string]string{
		"status":    "pending",
		"timestamp": time.Now().Format(time.RFC3339),
	}); err != nil {
		return "", fmt.Errorf("failed to update order status: %v", err)
	}

	_ = qs.rdb.Expire(ctx, key, 10*time.Minute)
	return orderID, nil
}
