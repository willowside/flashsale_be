package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
	PostgresDBName   string
	PostgresSSLMode  string
	ServerPort       string
	RedisHost        string
	// PostgresDSN      string
	RedisPort     string
	RedisPassword string
	MQUrl         string
}

func LoadConfig() *Config {
	godotenv.Load()

	cfg := &Config{
		ServerPort:       getEnv("SERVER_PORT", "8080"),
		PostgresHost:     getEnv("POSTGRES_HOST", "postgres"),
		PostgresPort:     getEnv("POSTGRES_PORT", "5432"),
		PostgresUser:     getEnv("POSTGRES_USER", "user"),
		PostgresPassword: getEnv("POSTGRES_PASSWORD", "password"),
		PostgresDBName:   getEnv("POSTGRES_DB", "flashsale"),
		PostgresSSLMode:  getEnv("POSTGRES_SSLMODE", "disable"),
		RedisHost:        getEnv("REDIS_HOST", "redis"),
		RedisPort:        getEnv("REDIS_PORT", "6379"),
		RedisPassword:    getEnv("REDIS_PASSWORD", ""),
		MQUrl:            getEnv("MQ_URL", "amqp://guest:guest@rabbitmq:5672/"),
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
