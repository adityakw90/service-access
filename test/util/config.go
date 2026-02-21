package testutil

import (
	"strconv"
	"testing"
	"time"

	"github.com/adityakw90/service-access/internal/config"
)

func LoadTestConfig(t *testing.T) (*config.Config, error) {
	t.Helper()

	dbPort, err := strconv.Atoi(getEnv("DATABASE_PORT", "5432"))
	if err != nil {
		return nil, err
	}
	redisPort, err := strconv.Atoi(getEnv("REDIS_PORT", "6379"))
	if err != nil {
		return nil, err
	}
	rabbitPort, err := strconv.Atoi(getEnv("RABBITMQ_PORT", "5672"))
	if err != nil {
		return nil, err
	}
	return &config.Config{
		Database: config.DatabaseConfig{
			Host:                  getEnv("DATABASE_HOST", "localhost"),
			Port:                  dbPort,
			User:                  getEnv("DATABASE_USER", "postgres"),
			Password:              getEnv("DATABASE_PASSWORD", "postgres"),
			Name:                  getEnv("DATABASE_NAME", "test_service_access"),
			SslMode:               getEnv("DATABASE_SSL_MODE", "disable"),
			Timezone:              "UTC",
			MinConns:              1,
			MinIdleConns:          1,
			MaxConns:              10,
			MaxConnIdleTime:       10 * time.Minute,
			MaxConnLifetime:       30 * time.Minute,
			MaxConnLifetimeJitter: 5 * time.Minute,
			HealthCheckPeriod:     1 * time.Minute,
		},
		Redis: config.RedisConfig{
			Host:              getEnv("REDIS_HOST", "localhost"),
			Port:              redisPort,
			User:              getEnv("REDIS_USER", ""),
			Password:          getEnv("REDIS_PASSWORD", ""),
			DB:                0,
			PoolSize:          10,
			PoolTimeout:       5 * time.Second,
			ConnectionIdleMin: 1,
		},
		Rabbit: config.RabbitConfig{
			Host:                 getEnv("RABBITMQ_HOST", "localhost"),
			Port:                 rabbitPort,
			User:                 getEnv("RABBITMQ_USER", "rabbit"),
			Password:             getEnv("RABBITMQ_PASSWORD", "password"),
			Vhost:                getEnv("RABBITMQ_VHOST", "/"),
			ReconnectInterval:    1 * time.Second,
			ReconnectMaxAttempts: 0, // 0 means infinite retries
		},
		Observer: config.ObserverConfig{
			Auth:     true,
			User:     true,
			Device:   true,
			UserFile: true,
			Pin:      true,
		},
		EventPublisher: config.EventPublisherConfig{
			HTTP: config.PublisherHTTPConfig{
				Enabled: true,
				URL:     "http://localhost:8080",
				Timeout: 5 * time.Second,
			},
			Redis: config.PublisherRedisConfig{
				Enabled: true,
				Name:    "test_service_access",
			},
			RabbitMQ: config.PublisherRabbitMQConfig{
				Enabled:          true,
				Exchange:         "test_service_access",
				ExchangeType:     "topic",
				RoutingKeyPrefix: "test_service_access.",
				Durable:          true,
			},
		},
	}, nil
}
