package queue

import (
	"context"
	"encoding/json"
	"flashsale/internal/dto"
	"flashsale/pkg/mq"
	"fmt"
)

type RabbitMQDLQPublisher struct {
	Client *mq.RabbitMQClient
}

func NewRabbitMQDLQPublisher(client *mq.RabbitMQClient) *RabbitMQDLQPublisher {
	return &RabbitMQDLQPublisher{Client: client}
}

func (p *RabbitMQDLQPublisher) Publish(ctx context.Context, msg dto.DLQMessage) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal dlq msg failed: %w", err)
	}

	return p.Client.Publish(ctx, body)
}
