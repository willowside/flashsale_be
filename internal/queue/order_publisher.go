package queue

import (
	"context"
	"encoding/json"
	"flashsale/internal/domain"
	"flashsale/internal/service/serviceiface"
	"flashsale/pkg/mq"
	"fmt"
	"time"
)

type RabbitMQOrderPublisher struct {
	Client *mq.RabbitMQClient
}

func NewRabbitMQOrderPublisher(client *mq.RabbitMQClient) serviceiface.OrderPublisher {
	return &RabbitMQOrderPublisher{Client: client}
}

func (p *RabbitMQOrderPublisher) PublishOrder(ctx context.Context, orderID, userID, productID string) error {
	// 1. prepare msg content
	msg := domain.OrderMessage{
		OrderID:   orderID,
		UserID:    userID,
		ProductID: productID,
		Timestamp: NowUnix(),
	}
	// 2. msg content to json format
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal order msg: %w", err)
	}

	return p.Client.Publish(ctx, body)
}

// time util
func NowUnix() int64 {
	return time.Now().Unix()
}
