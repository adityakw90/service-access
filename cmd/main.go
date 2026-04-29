package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpcAdapter "github.com/adityakw90/service-access/internal/adapter/api/grpc"
	"github.com/adityakw90/service-access/internal/adapter/event"
	"github.com/adityakw90/service-access/internal/adapter/executor"
	"github.com/adityakw90/service-access/internal/adapter/observer"
	"github.com/adityakw90/service-access/internal/adapter/repository"
	"github.com/adityakw90/service-access/internal/adapter/resolver"
	"github.com/adityakw90/service-access/internal/adapter/security"
	"github.com/adityakw90/service-access/internal/config"
	portEvent "github.com/adityakw90/service-access/internal/core/port/event"
	"github.com/adityakw90/service-access/internal/core/service"
	"github.com/adityakw90/service-access/internal/infra"
)

func main() {
	// Handle --version flag early (before loading config)
	handleVersionFlag()

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
	logger.Info("connected to database", map[string]interface{}{
		"host": cfg.Database.Host,
		"port": cfg.Database.Port,
	})

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
	logger.Info("connected to redis", map[string]interface{}{
		"host": cfg.Redis.Host,
		"port": cfg.Redis.Port,
	})

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
		var backendsPublisher []portEvent.EventPublisher

		// setup http publisher
		if cfg.EventPublisher.HTTP.Enabled {
			eventHttpPublisher := event.NewHTTPPublisher(
				cfg.EventPublisher.HTTP.URL,
				cfg.EventPublisher.HTTP.Timeout,
				iMon.Logger,
				iMon.Tracer,
			)
			backendsPublisher = append(backendsPublisher, eventHttpPublisher)
		}

		// setup rabbitmq publisher
		if cfg.EventPublisher.RabbitMQ.Enabled {
			var rabbitmqConn *infra.Rabbit
			rabbitmqConn, err = infra.NewRabbitConnection(ctx, infra.RabbitConfig{
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
				"host":  cfg.Rabbit.Host,
				"port":  cfg.Rabbit.Port,
				"user":  cfg.Rabbit.User,
				"vhost": cfg.Rabbit.Vhost,
			})

			eventRabbitPublisher := event.NewRabbitmqPublisher(
				rabbitmqConn,
				event.RabbitmqPublisherConfig{
					Exchange:         cfg.EventPublisher.RabbitMQ.Exchange,
					RoutingKeyPrefix: cfg.EventPublisher.RabbitMQ.RoutingKeyPrefix,
					ConfirmTimeout:   cfg.EventPublisher.RabbitMQ.ConfirmTimeout,
					MaxRetries:       cfg.EventPublisher.RabbitMQ.MaxRetries,
					RetryInterval:    cfg.EventPublisher.RabbitMQ.RetryInterval,
				},
				iMon.Logger,
				iMon.Tracer,
			)
			backendsPublisher = append(backendsPublisher, eventRabbitPublisher)
		}

		if len(backendsPublisher) == 1 {
			eventPublisher = backendsPublisher[0]
		} else if len(backendsPublisher) > 1 {
			eventPublisher = event.NewMultiEventPublisher(
				iMon.Logger,
				iMon.Tracer,
				backendsPublisher...,
			)
		}
	}

	// Default to no-op if no publishers configured
	if eventPublisher == nil {
		eventPublisher = event.NewNoOpPublisher()
	}
	// Close publisher on shutdown
	defer eventPublisher.Close()

	// create executor
	exc := executor.NewServiceExecutor(iMon.Logger, iMon.Tracer)

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
		exc,
		permissionObserver,
	)
	groupService := service.NewGroupService(
		uow,
		repoProvider,
		eventPublisher,
		uidGen,
		resolverProvider,
		exc,
		groupObserver,
		groupPermissionObserver,
	)
	roleService := service.NewRoleService(
		uow,
		repoProvider,
		eventPublisher,
		uidGen,
		resolverProvider,
		exc,
		roleObserver,
		rolePermissionObserver,
	)
	subjectService := service.NewSubjectService(
		uow,
		repoProvider,
		eventPublisher,
		exc,
		subjectObserver,
	)
	accessService := service.NewAccessService(
		repoProvider,
		eventPublisher,
		exc,
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
