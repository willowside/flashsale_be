package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort    string
	RedisHost     string
	PostgresDSN   string
	RedisPort     string
	RedisPassword string
	MQUrl         string
}

func LoadConfig() *Config {
	godotenv.Load()

	cfg := &Config{
		ServerPort:    getEnv("SERVER_PORT", "8080"),
		PostgresDSN:   getEnv("POSTGRES_DSN", "postgres://user:pass@localhost:5432/flashsale?sslmode=disable"),
		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		MQUrl:         getEnv("MQ_URL", "amqp://guest:guest@localhost:5672/"),
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
