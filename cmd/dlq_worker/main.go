package main

import (
	"context"
	"encoding/json"
	"flashsale/internal/cache"
	"flashsale/internal/dto"
	"flashsale/internal/repository"
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
	compensator := service.NewOrderCompensator(repo)
	dlqWorker := worker.NewDLQWorker(compensator)
	log.Println("DLQ worker started")

	for d := range consumer {
		var msg dto.DLQMessage
		if err := json.Unmarshal(d.Body, &msg); err != nil {
			log.Println("[DLQ] invalid message:", err)
			_ = d.Ack(false)
			continue
		}

		dlqWorker.Handle(context.Background(), msg)
		_ = d.Ack(false)
	}
}
