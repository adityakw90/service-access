package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpcAdapter "github.com/adityakw90/service-access/internal/adapter/api/grpc"
	"github.com/adityakw90/service-access/internal/adapter/observer"
	"github.com/adityakw90/service-access/internal/adapter/publisher"
	"github.com/adityakw90/service-access/internal/adapter/repository"
	"github.com/adityakw90/service-access/internal/adapter/resolver"
	"github.com/adityakw90/service-access/internal/adapter/security"
	"github.com/adityakw90/service-access/internal/config"
	portEvent "github.com/adityakw90/service-access/internal/core/port/event"
	"github.com/adityakw90/service-access/internal/core/service"
	"github.com/adityakw90/service-access/internal/infra"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	// Initialize logger
	logger := infra.NewLogger()

	// initialize monitoring
	iMon, err := infra.NewMonitoring(&infra.MonitoringConfig{
		ServiceName:        cfg.Monitoring.ServiceName,
		Environment:        cfg.Monitoring.Environment,
		InstanceName:       cfg.Instance.Name,
		InstanceHost:       cfg.Instance.Host,
		LoggerLevel:        cfg.Monitoring.Logger.Level,
		TracerProvider:     cfg.Monitoring.Tracer.Provider,
		TracerProviderHost: cfg.Monitoring.Tracer.ProviderHost,
		TracerProviderPort: cfg.Monitoring.Tracer.ProviderPort,
		TracerSampleRatio:  cfg.Monitoring.Tracer.SampleRatio,
		TracerInsecure:     cfg.Monitoring.Tracer.Insecure,
		MetricProvider:     cfg.Monitoring.Metric.Provider,
		MetricProviderHost: cfg.Monitoring.Metric.ProviderHost,
		MetricProviderPort: cfg.Monitoring.Metric.ProviderPort,
		MetricInsecure:     cfg.Monitoring.Metric.Insecure,
	})
	if err != nil {
		logger.Fatal("failed to initialize monitoring", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// start context
	ctx := context.Background()

	// Connect to PostgreSQL
	dbPool, err := infra.NewPostgreConnection(ctx, &infra.PostgreConfig{
		Host:                  cfg.Database.Host,
		Port:                  cfg.Database.Port,
		User:                  cfg.Database.User,
		Password:              cfg.Database.Password,
		Name:                  cfg.Database.Name,
		SslMode:               cfg.Database.SslMode,
		Timezone:              cfg.Database.Timezone,
		MinConns:              cfg.Database.MinConns,
		MinIdleConns:          cfg.Database.MinIdleConns,
		MaxConns:              cfg.Database.MaxConns,
		MaxConnIdleTime:       cfg.Database.MaxConnIdleTime,
		MaxConnLifetime:       cfg.Database.MaxConnLifetime,
		MaxConnLifetimeJitter: cfg.Database.MaxConnLifetimeJitter,
		HealthCheckPeriod:     cfg.Database.HealthCheckPeriod,
		QueryExecMode:         cfg.Database.QueryExecMode,
	})
	if err != nil {
		logger.Fatal("failed to connect to database", map[string]interface{}{
			"error": err.Error(),
		})
	}
	defer dbPool.Close()
	logger.Info("connected to database", nil)

	// Connect to Redis using infra layer
	redisClient, err := infra.NewRedisConnection(context.Background(), &infra.RedisConfig{
		Host:              cfg.Redis.Host,
		Port:              cfg.Redis.Port,
		User:              cfg.Redis.User,
		Password:          cfg.Redis.Password,
		DB:                cfg.Redis.DB,
		PoolSize:          cfg.Redis.PoolSize,
		PoolTimeout:       cfg.Redis.PoolTimeout,
		ConnectionIdleMin: cfg.Redis.ConnectionIdleMin,
	})
	if err != nil {
		logger.Fatal("failed to connect to redis", map[string]interface{}{
			"error": err.Error(),
		})
	}
	defer redisClient.Close()
	logger.Info("connected to redis", nil)

	// Connect to RabbitMQ using infra layer (if enabled)
	var rabbitmqConn *infra.RabbitMQConnection
	if cfg.EventPublisher.RabbitMQ.Enabled {
		rabbitmqConn, err = infra.NewRabbitMQConnection(ctx, infra.RabbitMQConfig{
			Host:                 cfg.Rabbit.Host,
			Port:                 cfg.Rabbit.Port,
			User:                 cfg.Rabbit.User,
			Password:             cfg.Rabbit.Password,
			Vhost:                cfg.Rabbit.Vhost,
			ReconnectInterval:    cfg.Rabbit.ReconnectInterval,
			ReconnectMaxAttempts: cfg.Rabbit.ReconnectMaxAttempts,
		}, iMon.Logger)
		if err != nil {
			logger.Fatal("failed to connect to rabbitmq", map[string]interface{}{
				"error": err.Error(),
			})
		}
		defer rabbitmqConn.Close()
		logger.Info("connected to rabbitmq", map[string]interface{}{
			"exchange": cfg.EventPublisher.RabbitMQ.Exchange,
		})
	}

	// Connect to Kafka using infra layer (if enabled)
	var kafkaConn *infra.KafkaConnection
	if cfg.EventPublisher.Kafka.Enabled {
		kafkaConn, err = infra.NewKafkaConnection(ctx, infra.KafkaConfig{
			Brokers:              cfg.Kafka.Brokers,
			MaxMessageBytes:      cfg.Kafka.MaxMessageBytes,
			Timeout:              time.Duration(cfg.Kafka.TimeoutSeconds) * time.Second,
			Compression:          cfg.Kafka.Compression,
			ReconnectInterval:    1 * time.Second,
			ReconnectMaxAttempts: 0, // 0 means infinite retries
		}, iMon.Logger)
		if err != nil {
			logger.Fatal("failed to connect to kafka", map[string]interface{}{
				"error": err.Error(),
			})
		}
		defer kafkaConn.Close()
		logger.Info("connected to kafka", map[string]interface{}{
			"brokers": cfg.Kafka.Brokers,
			"topic":   cfg.EventPublisher.Kafka.Topic,
		})
	}

	// Initialize repositories
	repoProvider := repository.NewRepositoryProvider(dbPool)
	uow := repository.NewUnitOfWork(dbPool)

	// initialzie resolver
	resolverProvider := resolver.NewResolverProvider(
		dbPool,
		redisClient,
		cfg.App.Code+":resolver",
		1*time.Hour,
		iMon.Logger,
		iMon.Tracer,
	)

	// --- Event Publishers Setup ---
	var eventPublisher portEvent.EventPublisher
	if cfg.EventPublisher.Enabled {
		var backends []portEvent.EventPublisher

		// Redis Stream backend for all events
		if cfg.EventPublisher.Redis.Enabled && redisClient != nil {
			redisStreamPub, err := publisher.NewRedisPublisher(
				redisClient,
				publisher.RedisPublisherConfig{
					Stream: cfg.EventPublisher.Redis.Name,
					MaxLen: cfg.EventPublisher.Redis.MaxLen,
					Source: cfg.Instance.Name,
				},
				logger,
			)
			if err != nil {
				logger.Fatal("failed to create redis stream publisher", map[string]interface{}{
					"error": err.Error(),
				})
			}
			backends = append(backends, redisStreamPub)
		}

		// Kafka backend for all events
		// Use the existing connection if available (created earlier in main.go)
		if cfg.EventPublisher.Kafka.Enabled {
			if kafkaConn == nil {
				logger.Fatal("kafka connection is nil but kafka is enabled", nil)
			}
			kafkaPub := publisher.NewKafkaPublisherWithConn(
				kafkaConn,
				cfg.EventPublisher.Kafka.Topic,
				cfg.Instance.Name,
			)
			backends = append(backends, kafkaPub)
		}

		// RabbitMQ backend for all events
		// Use the existing connection if available (created earlier in main.go)
		if cfg.EventPublisher.RabbitMQ.Enabled {
			if rabbitmqConn == nil {
				logger.Fatal("rabbitmq connection is nil but rabbitmq is enabled", nil)
			}
			rabbitPub := publisher.NewRabbitMQPublisher(
				rabbitmqConn,
				publisher.RabbitMQPublisherConfig{
					Source:           cfg.Instance.Name,
					Exchange:         cfg.EventPublisher.RabbitMQ.Exchange,
					ExchangeType:     cfg.EventPublisher.RabbitMQ.ExchangeType,
					RoutingKeyPrefix: cfg.EventPublisher.RabbitMQ.RoutingKeyPrefix,
					Durable:          cfg.EventPublisher.RabbitMQ.Durable,
					ConfirmTimeout:   cfg.EventPublisher.RabbitMQ.ConfirmTimeout,
					MaxRetries:       cfg.EventPublisher.RabbitMQ.MaxRetries,
					RetryInterval:    cfg.EventPublisher.RabbitMQ.RetryInterval,
					QueueName:        cfg.EventPublisher.RabbitMQ.QueueName,
					QueueDurable:     cfg.EventPublisher.RabbitMQ.QueueDurable,
					QueueAutoDelete:  cfg.EventPublisher.RabbitMQ.QueueAutoDelete,
					QueueExclusive:   cfg.EventPublisher.RabbitMQ.QueueExclusive,
					QueueEnabled:     cfg.EventPublisher.RabbitMQ.QueueEnabled,
				},
			)
			// Setup infrastructure (exchange and optionally queue)
			if err := rabbitPub.SetupInfrastructure(); err != nil {
				logger.Fatal("failed to setup rabbitmq infrastructure", map[string]interface{}{
					"error":    err.Error(),
					"exchange": cfg.EventPublisher.RabbitMQ.Exchange,
				})
			}
			backends = append(backends, rabbitPub)
		}

		// HTTP backend for all events
		if cfg.EventPublisher.HTTP.Enabled {
			backends = append(backends, publisher.NewHTTPPublisher(
				publisher.HttpPublisherConfig{
					Endpoint: cfg.EventPublisher.HTTP.URL,
					Source:   cfg.Instance.Name,
					Timeout:  cfg.EventPublisher.HTTP.Timeout,
				},
			))
		}

		// Combine backends and wrap with async
		if len(backends) > 0 {
			multiBackend, err := publisher.NewMultiPublisher(logger, backends...)
			if err != nil {
				logger.Fatal("failed to create multi publisher", map[string]interface{}{
					"error": err.Error(),
				})
			}
			eventPublisher = publisher.NewAsyncPublisher(multiBackend, publisher.AsyncPublisherConfig{
				WorkerCount:   cfg.EventPublisher.WorkerCount,
				QueueSize:     cfg.EventPublisher.QueueSize,
				BatchSize:     cfg.EventPublisher.BatchSize,
				BatchTimeout:  cfg.EventPublisher.BatchTimeout,
				MaxRetries:    cfg.EventPublisher.MaxRetries,
				RetryInterval: cfg.EventPublisher.RetryInterval,
			})
		}
	}

	// Default to no-op if no publishers configured
	if eventPublisher == nil {
		eventPublisher = publisher.NewNoOpPublisher()
	}

	// Close publisher on shutdown
	defer eventPublisher.Close()

	// initialize observer
	// TODO: change to real observer after implemented
	permissionObserver := observer.NewPermissionObserver(iMon.Logger, iMon.Tracer)
	groupObserver := observer.NewGroupObserver(iMon.Logger, iMon.Tracer)
	groupPermissionObserver := observer.NewGroupPermissionObserver(iMon.Logger, iMon.Tracer)
	roleObserver := observer.NewRoleObserver(iMon.Logger, iMon.Tracer)
	rolePermissionObserver := observer.NewRolePermissionObserver(iMon.Logger, iMon.Tracer)
	subjectObserver := observer.NewSubjectObserver(iMon.Logger, iMon.Tracer)
	accessObserver := observer.NewAccessObserver(iMon.Logger, iMon.Tracer)

	// initialize service
	uidGen := security.NewUIDGenerator()
	permissionService := service.NewPermissionService(
		uow,
		repoProvider,
		eventPublisher,
		uidGen,
		resolverProvider,
		permissionObserver,
	)
	groupService := service.NewGroupService(
		uow,
		repoProvider,
		eventPublisher,
		uidGen,
		resolverProvider,
		groupObserver,
		groupPermissionObserver,
	)
	roleService := service.NewRoleService(
		uow,
		repoProvider,
		eventPublisher,
		uidGen,
		resolverProvider,
		roleObserver,
		rolePermissionObserver,
	)
	subjectService := service.NewSubjectService(
		uow,
		repoProvider,
		eventPublisher,
		subjectObserver,
	)
	accessService := service.NewAccessService(
		repoProvider,
		eventPublisher,
		accessObserver,
	)

	// setup gRpc Server
	srv := grpcAdapter.NewServer(
		permissionService,
		roleService,
		groupService,
		accessService,
		subjectService,
		iMon,
	)
	addr := fmt.Sprintf("%s:%d", cfg.App.IP, cfg.App.Port)
	// Handle shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		logger.Info("shutting down server", nil)
		srv.Stop()
	}()

	if err := srv.Start(addr); err != nil {
		logger.Fatal("failed to serve", map[string]interface{}{
			"error": err.Error(),
		})
	}
}
