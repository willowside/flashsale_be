package queue

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Manage connections, Cannel, QueueDeclare

// RabbitMQClient manage single queue connection/ channel
type RabbitMQClient struct {
	url       string
	queueName string

	conn    *amqp.Connection
	channel *amqp.Channel

	mu sync.RWMutex
}

// NewRabbitMQClient build client -> connection, & declare queue
// do retry connection, return error if all failed
func NewRabbitMQClient(url, queueName string) (*RabbitMQClient, error) {
	c := &RabbitMQClient{
		url:       url,
		queueName: queueName,
	}

	if err := c.connectWithRetry(5, 2*time.Second); err != nil {
		return nil, fmt.Errorf("rabbitmq connec failed: %w", err)
	}

	return c, nil

}

func (c *RabbitMQClient) connectWithRetry(attempts int, delay time.Duration) error {
	var lastErr error
	for i := 0; i < attempts; i++ {
		if err := c.connect(); err == nil {
			return nil
		} else {
			lastErr = err
			log.Printf("[rabbitMQ] connect attempt %d failed: %v - retry in %s", i+1, err, delay)
			time.Sleep(delay)
		}
	}
	return lastErr
}

// only do once connect & queue declaration
func (c *RabbitMQClient) connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	conn, err := amqp.Dial(c.url)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("open channel: %w", err)
	}

	_, err = ch.QueueDeclare(
		c.queueName,
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,   // args
	)

	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return fmt.Errorf("queue declare : %w", err)
	}

	c.conn = conn
	c.channel = ch
	return nil
}

// Close channel & conn
func (c *RabbitMQClient) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.channel != nil {
		_ = c.channel.Close()
		c.channel = nil
	}

	if c.conn != nil {
		_ = c.conn.Close()
		c.conn = nil
	}
}

// Publish: deleiver message to queue synchronously (blocking)
// body should be a serialized []byte, eg. JSON
func (c *RabbitMQClient) Publish(ctx context.Context, body []byte) error {
	c.mu.RLock()
	ch := c.channel
	c.mu.RUnlock()

	if c == nil {
		return errors.New("rabbitmq channel not available")
	}

	publishing := amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
		Timestamp:   time.Now(),
	}

	// use PublishWithContext to support context cancel
	if err := ch.PublishWithContext(ctx, "", c.queueName, false, false, publishing); err != nil {
		// if failed, one retry within a short period of time, then return error
		log.Printf("[rabbitmq] publish error: %v", err)
		go func() {
			// background reconnect attempt
			_ = c.connectWithRetry(3, 1*time.Second)
		}()
		return fmt.Errorf("publish failed: %w", err)
	}
	return nil
}

// Consume return an amqp.Delivery channel
func (c *RabbitMQClient) Consume(ctx context.Context, consumerName string, autoAck bool) (<-chan amqp.Delivery, error) {
	c.mu.RLock()
	ch := c.channel
	queue := c.queueName
	c.mu.RUnlock()

	if ch == nil {
		return nil, errors.New("rabbitmq channel not available")
	}

	msgs, err := ch.Consume(
		queue,
		consumerName,
		autoAck,
		false, // exclusive
		false, // noLocal, rebbitMQ not support
		false, // noWait
		nil,
	)

	if err != nil {
		return nil, fmt.Errorf("consume register failed: %w", err)
	}

	out := make(chan amqp.Delivery)

	// transfer amqp deliveries to out
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case d, ok := <-msgs:
				if !ok {
					// msgs channel closed -> try reconnect & rebuild consumer
					log.Println("[rabbitmq] deliveries channel closed, attempting reconnect")
					// try reconnect (blocking short time) & re register
					if err := c.connectWithRetry(5, 500*time.Millisecond); err != nil {
						log.Printf("[rabbitmq] reconnect failed: %v", err)
						return
					}
					// re register consumer
					c.mu.RLock()
					ch2 := c.channel
					c.mu.RUnlock()
					if ch2 == nil {
						return
					}
					msgs2, err := ch2.Consume(queue, consumerName, autoAck, false, false, false, nil)
					if err != nil {
						log.Printf("[rabbitmq] re-register consumer failed: %v", err)
						return
					}
					msgs = msgs2
					continue
				}
				// send to out
				select {
				case <-ctx.Done():
					return
				case out <- d:
				}
			}
		}
	}()
	return out, nil

}
