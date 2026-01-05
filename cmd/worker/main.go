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

			var orderMsg dto.OrderMessage
			processErr := service.Retry(3, func() error {
				// create a fresh timeout per attempt
				jobCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				return orderProcessor.ProcessOrder(jobCtx, d.Body)
			})

			if processErr != nil {
				// parse order message for DLQ
				if unmarshalErr := json.Unmarshal(d.Body, &orderMsg); unmarshalErr != nil {
					log.Printf("[Worker] Unmarshal failed: %v, message dropped", unmarshalErr)
					_ = d.Ack(false)
					continue
				}

				switch processErr {
				case worker.ErrOutOfStock, worker.ErrLuaReject:
					// business logic failure -> order already marked FAILED inside processor
					// should have done in orderProcessor, no compensation needed so Ack directly
					log.Printf("[Worker] Business logic rejected: %v", processErr)
					_ = d.Ack(false)

				default:
					// system failure or force-fail
					log.Printf("[DLQ] order processing failed, sending to DLQ: %v", processErr)

					dlqMsg := dto.DLQMessage{
						OrderNo: orderMsg.OrderID,
						Reason:  processErr.Error(),
						Payload: dto.QueueOrderReq{
							OrderNo:   orderMsg.OrderID,
							UserID:    orderMsg.UserID,
							ProductID: mustParseProductID(orderMsg.ProductID),
						},
					}
					if err := dlqPublisher.Publish(ctx, dlqMsg); err != nil {
						log.Printf("[DLQ Critical] Failed to publish to DLQ: %v", err)
						// if DLQ failed as well, No Ack, let  message retry in queue (Unacked)
						_ = d.Nack(false, true)
					} else {
						log.Printf("[DLQ] Success published Order: %s", orderMsg.OrderID)
						_ = d.Ack(false)
					}
				}
				continue
			}
			// Success
			log.Printf("[Worker] Order processed successfully: %s", orderMsg.OrderID)
			_ = d.Ack(false)
		}
	}
}

func mustParseProductID(s string) int64 {
	id, _ := strconv.ParseInt(s, 10, 64)
	return id
}
