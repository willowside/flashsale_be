package main

import (
	"context"
	"encoding/json"
	"flashsale/internal/cache"
	"flashsale/internal/dto"
	"flashsale/internal/repository"
	"flashsale/internal/repository/redis"
	"flashsale/internal/service"
	"flashsale/internal/worker"
	"flashsale/pkg/config"
	"flashsale/pkg/db"
	"flashsale/pkg/mq"
	"log"
)

func main() {
	cfg := config.LoadConfig()
	dlqQueue := "flashsale_order_dlq"

	// infra
	_ = db.InitPostgresDB(
		cfg.PostgresHost,
		cfg.PostgresPort,
		cfg.PostgresUser,
		cfg.PostgresPassword,
		cfg.PostgresDBName,
		cfg.PostgresSSLMode,
	)
	_ = cache.InitRedis(cfg.RedisHost, cfg.RedisPort, "", 0)

	mqClient, err := mq.NewRabbitMQClient(cfg.MQUrl, dlqQueue)
	if err != nil {
		log.Fatal(err)
	}
	defer mqClient.Close()

	consumer, err := mqClient.Consume(context.Background(), "dlq-worker", false)
	if err != nil {
		log.Fatal(err)
	}

	// dependencies
	repo := repository.NewOrderRepository(db.Pool, "postgres")
	redisStockRepo := redis.NewRedisStockRepo(cache.Rdb)
	compensator := service.NewOrderCompensator(repo, redisStockRepo)
	dlqWorker := worker.NewDLQWorker(compensator)
	log.Println("DLQ worker started")

	for d := range consumer {
		log.Printf("[DLQ] message received: %s", string(d.Body))
		var msg dto.DLQMessage
		if err := json.Unmarshal(d.Body, &msg); err != nil {
			log.Println("[DLQ] invalid message:", err)
			_ = d.Ack(false) // Ack when parse error
			continue
		}

		// err := compensator.

		err := dlqWorker.Handle(context.Background(), msg)
		if err != nil {
			log.Printf("[DLQ] compensation failed: %v", err)
			// compensation failed, Nack to requeue(or logging)
			_ = d.Nack(false, true)
		} else {
			_ = d.Ack(false) // Ack when compensation success
			log.Panicln("[DLQ] compensation success, Ack")
		}

	}
}
