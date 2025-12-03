package queue

import (
	"context"
	"encoding/json"
	"flashsale/internal/service"
	"fmt"
	"time"
)

type OrderMessage struct {
	UserID    string `json:"user_id"`
	ProductID string `json:"product_id"`
	Timestamp int64  `json:"timestamp"`
}

type RabbitMQOrderPublisher struct {
	client *RabbitMQClient
}

func NewRabbitMQOrderPublisher(client *RabbitMQClient) service.OrderPublisher {
	return &RabbitMQOrderPublisher{client: client}
}

func (p *RabbitMQOrderPublisher) PublishOrder(ctx context.Context, userID, productID string) error {
	// 1. prepare msg content
	msg := OrderMessage{
		UserID:    userID,
		ProductID: productID,
		Timestamp: NowUnix(),
	}
	// 2. msg content to json format
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal order msg: %w", err)
	}

	return p.client.Publish(ctx, body)

	// // 3. write into RabbitMQ
	// err = p.client.GetChannel().PublishWithContext(
	// 	ctx,
	// 	"",
	// 	p.client.queueName,
	// 	false,
	// 	false,
	// 	amqp091.Publishing{
	// 		ContentType: "application/json",
	// 		Body:        body,
	// 	},
	// )
	// if err != nil {
	// 	fmt.Errorf("failed to publish order: %w", err)
	// }
	// return nil
}

// time util
func NowUnix() int64 {
	return time.Now().Unix()
}
