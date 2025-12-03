package main

import (
	"context"
	"flashsale/internal/cache"
	"flashsale/internal/queue"
	"flashsale/internal/worker"
	"flashsale/pkg/config"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := config.LoadConfig()
	queueName := "flashsale_order_queue"

	// init redis
	if err := cache.InitRedis(cfg.RedisHost, cfg.RedisPort, "", 0); err != nil {
		log.Fatalf("redis init failed: %v", err)
	}
	// load lua scripts
	luaScripts, err := cache.LoadLuaScripts(cache.Rdb, "./scripts")
	if err != nil {
		log.Fatalf("load lua failed: %v", err)
	}
	cache.SetLuaScripts(luaScripts)

	// Init RabbitMQ
	mqClient, err := queue.NewRabbitMQClient(
		cfg.MQUrl,
		queueName,
	)
	if err != nil {
		log.Fatalf("RabbitMQ init failed: %v", err)
	}
	defer mqClient.Close()
	log.Println("Worker Started. Waiting for messages...")

	consumer, err := mqClient.Consume(context.Background(), "worker-1", false) // false, manual ack
	if err != nil {
		log.Fatalf("consume failed: %v", err)
	}

	orderProcessor := worker.NewOrderProcessor(cache.Rdb, luaScripts)
	log.Println("worker started, awaiting messages...")

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
			if err := orderProcessor.ProcessOrder(ctx, d.Body); err != nil {
				log.Printf("process failed: %v, nack and requeue", err)
				_ = d.Nack(false, true)
				continue
			}
			_ = d.Ack(false)
		}
	}
}
