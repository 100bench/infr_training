package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	// Database
	DSN string

	// Cache
	RedisAddr string
	CacheSize int
	Ttl time.Duration
	// Kafka
	KafkaBrokers []string
	KafkaTopic   string

	// Relay
	RelayBatchSize    int
	RelayPollInterval time.Duration

	// Notify service (пока оставь, потом уберёшь)
	NotifyServiceAddr string

	// HTTP
	ServerPort string
}

func Load() (*Config, error) {
	cfg := &Config{
		DSN:       getEnv("DSN_STRING", "postgres://user:password@db:5432/todo_store?sslmode=disable"),
		RedisAddr: getEnv("REDIS_ADDR", ""),

		CacheSize: getEnvInt("CACHE_SIZE", 1000),
		Ttl: getEnvDuration("CACHE_TTL", 10*time.Minute),

		KafkaBrokers: []string{getEnv("KAFKA_BROKERS", "localhost:9092")},
		KafkaTopic:   getEnv("KAFKA_TOPIC", "task.events"),

		RelayBatchSize:    getEnvInt("RELAY_BATCH_SIZE", 100),
		RelayPollInterval: getEnvDuration("RELAY_POLL_INTERVAL", 500*time.Millisecond),

		NotifyServiceAddr: getEnv("NOTIFY_SERVICE_ADDR", ""),
		ServerPort:        getEnv("SERVER_PORT", "8080"),
	}

	// Валидация обязательных полей
	if cfg.DSN == "" {
		return nil, fmt.Errorf("DSN_STRING is required")
	}
	if cfg.RedisAddr == "" {
		return nil, fmt.Errorf("REDIS_ADDR is required")
	}

	return cfg, nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return defaultVal
}