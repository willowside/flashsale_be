package main

import (
	"flashsale/internal/cache"
	"flashsale/internal/handler"
	"flashsale/internal/queue"
	"flashsale/internal/repository"
	"flashsale/internal/router"
	"flashsale/internal/service"
	"flashsale/pkg/config"
	"flashsale/pkg/db"
	"flashsale/pkg/mq"
	"log"
)

func main() {

	cfg := config.LoadConfig()

	queueName := "flashsale_order_queue"
	ttlSeconds := 60

	// ------ init Postgres ------
	if err := db.InitPostgresDB(
		cfg.PostgresHost,
		cfg.PostgresPort,
		cfg.PostgresUser,
		cfg.PostgresPassword,
		cfg.PostgresDBName,
		cfg.PostgresSSLMode,
	); err != nil {
		log.Fatalf("Postgres init failed: %v", err)
	}

	// ------ init Redis ------
	err := cache.InitRedis(cfg.RedisHost, cfg.RedisPort, cfg.RedisPassword, 0)
	if err != nil {
		log.Fatalf("Redis init failed: %v", err)
	}

	// load precheck lua scripts
	scripts, err := cache.LoadLuaScripts(cache.Rdb, "./scripts")
	if err != nil {
		log.Fatal(err)
	}

	cache.SetLuaScripts(scripts)

	// ------ init RabbitMQ ------
	mqClient, err := mq.NewRabbitMQClient(
		cfg.MQUrl,
		queueName,
	)
	if err != nil {
		log.Fatalf("RabbitMQ init failed: %v", err)
	}
	defer mqClient.Close()

	orderRepo := repository.NewOrderRepository(db.Pool, "postgres")
	productRepo := repository.NewProductRepository(db.Pool, "postgres")

	// init OrderService + publisher
	orderPublisher := queue.NewRabbitMQOrderPublisher(mqClient)

	// init Service
	orderService := service.NewOrderService(orderPublisher, scripts, ttlSeconds)
	queueService := service.NewQueueService(cache.Rdb, orderPublisher)
	stockService := service.NewStockService(cache.Rdb, productRepo)
	resultService := service.NewOrderResultService(orderRepo)

	// init Router/Gin http server
	orderHandler := handler.NewOrderHandler(orderService)
	orderQueueHandler := handler.NewQueueHandler(queueService)
	stockHandler := handler.NewStockHandler(stockService)
	resultHandler := handler.NewOrderResultHandler(resultService)
	r := router.SetupRouter(
		orderHandler,
		orderQueueHandler,
		stockHandler,
		resultHandler,
	)

	log.Println("Flash Sale API Server running on : 8080")

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("API server failed: %v", err)
	}
}
