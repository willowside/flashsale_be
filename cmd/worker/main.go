package main

import (
	"context"
	"encoding/json"
	"flashsale/internal/cache"
	"flashsale/internal/dto"
	queue "flashsale/internal/mq"
	"flashsale/internal/repository"
	"flashsale/internal/service"
	"flashsale/internal/worker"
	"flashsale/pkg/config"
	"flashsale/pkg/db"
	"flashsale/pkg/mq"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func main() {
	cfg := config.LoadConfig()
	queueName := "flashsale_order_queue"
	dlqQueueName := "flashsale_order_dlq"

	// graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigc
		log.Println("shutdown signal received")
		cancel()
	}()

	// init Postgres
	if err := db.InitPostgresDB(cfg.PostgresHost, cfg.PostgresPort, cfg.PostgresUser, cfg.PostgresPassword, cfg.PostgresDBName, cfg.PostgresSSLMode); err != nil {
		log.Fatalf("Postgres init failed: %v", err)
	}
	// init redis
	if err := cache.InitRedis(cfg.RedisHost, cfg.RedisPort, "", 0); err != nil {
		log.Fatalf("redis init failed: %v", err)
	}
	// load lua scripts
	scripts, err := cache.LoadLuaScripts(cache.Rdb, "./scripts")
	if err != nil {
		log.Fatal(err)
	}

	cache.SetLuaScripts(scripts)

	// Init RabbitMQ
	mqClient, err := mq.NewRabbitMQClient(
		cfg.MQUrl,
		queueName,
	)
	if err != nil {
		log.Fatalf("RabbitMQ init failed: %v", err)
	}
	defer mqClient.Close()
	log.Println("Worker Started. Waiting for messages...")

	consumer, err := mqClient.Consume(ctx, "worker-1", false) // false, manual ack
	if err != nil {
		log.Fatalf("consume failed: %v", err)
	}
	// Init DLQ RabbitMQ
	dlqClient, err := mq.NewRabbitMQClient(
		cfg.MQUrl,
		dlqQueueName,
	)
	if err != nil {
		log.Fatalf("DLQ RabbitMQ init failed: %v", err)
	}
	defer dlqClient.Close()

	dlqPublisher := queue.NewRabbitMQDLQPublisher(dlqClient)

	repo := repository.NewOrderRepository(db.Pool, "postgres")
	orderProcessor := worker.NewOrderProcessor(repo, scripts)
	log.Println("worker started, awaiting messages...")

	for {
		select {
		case <-ctx.Done():
			log.Println("worker exiting")
			return

		case d, ok := <-consumer:
			if !ok {
				log.Println("consumer channel closed")
				return
			}

			err := service.Retry(3, func() error {
				// create a fresh timeout per attempt
				jobCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				return orderProcessor.ProcessOrder(jobCtx, d.Body)
			})

			if err != nil {
				switch err {
				case worker.ErrOutOfStock, worker.ErrLuaReject:
					// business logic failure -> order already marked FAILED inside processor
					_ = d.Ack(false)
					continue

				default:
					log.Printf("[DLQ] order processing failed: %v", err)

					var orderMsg dto.OrderMessage
					if err := json.Unmarshal(d.Body, &orderMsg); err != nil {
						log.Printf("[DLQ] failed to unmarshal order message: %v", err)
						_ = d.Ack(false)
						continue
					}

					_ = dlqPublisher.Publish(ctx, dto.DLQMessage{
						OrderNo: orderMsg.OrderID,
						Reason:  err.Error(),
						Payload: dto.QueueOrderReq{
							OrderNo:   orderMsg.OrderID,
							UserID:    orderMsg.UserID,
							ProductID: mustParseProductID(orderMsg.ProductID),
						},
					})
					_ = d.Ack(false)
					continue
				}
			}
			_ = d.Ack(false)
		}
	}
}

func mustParseProductID(s string) int64 {
	id, _ := strconv.ParseInt(s, 10, 64)
	return id
}
