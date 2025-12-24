package queue

import (
	"context"
	"encoding/json"
	"flashsale/internal/dto"
	"flashsale/pkg/mq"
	"fmt"
)

type RabbitMQDLQProducer struct {
	client *mq.RabbitMQClient
}

func NewRabbitMQDLQProducer(client *mq.RabbitMQClient) *RabbitMQDLQProducer {
	return &RabbitMQDLQProducer{client: client}
}

func (p *RabbitMQDLQProducer) Publish(
	ctx context.Context,
	msg dto.DLQMessage,
) error {

	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal dlq msg: %w", err)
	}

	return p.client.Publish(ctx, body)
}
