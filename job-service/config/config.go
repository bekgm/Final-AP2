package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	GRPCPort    string
	DatabaseURL string
	RedisURL    string
	RabbitMQURL string
	RabbitMQExchange string

	PostgresPingRetries  int
	PostgresPingInterval time.Duration

	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string
}

func Load() *Config {
	return &Config{
		GRPCPort:         getEnv("GRPC_PORT", ":50052"),
		DatabaseURL:      getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/jobdb?sslmode=disable"),
		RedisURL:         getEnv("REDIS_URL", "redis://localhost:6379"),
		RabbitMQURL:      getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		RabbitMQExchange: getEnv("RABBITMQ_EXCHANGE", "jobs"),

		PostgresPingRetries:  getEnvInt("POSTGRES_PING_RETRIES", 10),
		PostgresPingInterval: getEnvDurationSeconds("POSTGRES_PING_INTERVAL_SECONDS", 3),

		SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     getEnv("SMTP_PORT", "587"),
		SMTPUsername: getEnv("SMTP_USERNAME", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:     getEnv("SMTP_FROM", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func getEnvDurationSeconds(key string, fallbackSeconds int) time.Duration {
	seconds := getEnvInt(key, fallbackSeconds)
	if seconds < 0 {
		seconds = 0
	}
	return time.Duration(seconds) * time.Second
}
