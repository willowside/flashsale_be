package main

import (
	"flashsale/internal/cache"
	"flashsale/internal/handler"
	"flashsale/internal/queue"
	"flashsale/internal/router"
	"flashsale/internal/service"
	"flashsale/pkg/config"
	"log"
)

func main() {

	cfg := config.LoadConfig()

	queueName := "flashsale_order_queue"
	ttlSeconds := 60

	// _ = db.InitPostgres(cfg.PostgresDSN)
	// _ = redis.NewRedis(cfg.RedisAddr)
	// _ = mq.NewRabbitMQ(cfg.MQUrl)

	// r := gin.Default()

	// // health check
	// r.GET("/health", handler.HealthHandler)

	// log.Printf("🚀 API Server running on :%s\n", cfg.ServerPort)
	// r.Run(":" + cfg.ServerPort)

	// init Redis
	err := cache.InitRedis(cfg.RedisHost, cfg.RedisPort, cfg.RedisPassword, 0)
	if err != nil {
		log.Fatalf("Redis init failed: %v", err)
	}

	scripts, err := cache.LoadLuaScripts(cache.Rdb, "./scripts")
	if err != nil {
		log.Fatal(err)
	}

	cache.SetLuaScripts(scripts)

	// init RabbitMQ
	mqClient, err := queue.NewRabbitMQClient(
		cfg.MQUrl,
		queueName,
	)
	if err != nil {
		log.Fatalf("RabbitMQ init failed: %v", err)
	}
	defer mqClient.Close()

	// init OrderService + publisher
	orderPublisher := queue.NewRabbitMQOrderPublisher(mqClient)

	// init Service
	orderService := service.NewOrderService(orderPublisher, scripts, ttlSeconds)

	// init Router/Gin http server
	// r := gin.Default()
	h := handler.NewOrderHandler(orderService)
	r := router.SetupRouter(h)

	log.Println("Flash Sale API Server running on : 8080")

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("API server failed: %v", err)
	}

	// worker activate queue consumer (read queue -> call processor -> into DB)

}
