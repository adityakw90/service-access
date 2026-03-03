package testutil

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/adityakw90/go-monitoring"
	adapterobserver "github.com/adityakw90/service-access/internal/adapter/observer"
	"github.com/adityakw90/service-access/internal/adapter/publisher"
	"github.com/adityakw90/service-access/internal/adapter/repository"
	"github.com/adityakw90/service-access/internal/adapter/resolver"
	adaptersecurity "github.com/adityakw90/service-access/internal/adapter/security"
	"github.com/adityakw90/service-access/internal/config"
	"github.com/adityakw90/service-access/internal/core/domain/signal"
	portsecurity "github.com/adityakw90/service-access/internal/core/port/security"
	coreservice "github.com/adityakw90/service-access/internal/core/port/service"
	"github.com/adityakw90/service-access/internal/core/service"
	"github.com/adityakw90/service-access/internal/infra"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// TestServices holds all the services and dependencies for e2e testing.
type TestServices struct {
	Cfg               *config.Config
	DBPool            *pgxpool.Pool
	Redis             *redis.Client
	Monitoring        *monitoring.Monitoring
	UIDGenerator      portsecurity.UIDGenerator
	PermissionService coreservice.PermissionService
	GroupService      coreservice.GroupService
	RoleService       coreservice.RoleService
	SubjectService    coreservice.SubjectService
	AccessService     coreservice.AccessService
}

// SetupTestServices initializes all services for e2e testing.
// It connects to the database and Redis, initializes repositories,
// and creates service instances with test configuration.
func SetupTestServices(t *testing.T, ctx context.Context) (*TestServices, error) {
	t.Helper()

	// Load configuration
	cfg, err := LoadTestConfig(t)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize monitoring
	monitoring, err := infra.NewMonitoring(&infra.MonitoringConfig{
		ServiceName:         "test-service-access",
		Environment:         "development",
		InstanceName:        "test-instance",
		InstanceHost:        "test-host",
		LoggerLevel:         "error",
		LoggerCallerSkipNum: 1,
		TracerProvider:      "stdout",
		TracerProviderHost:  "localhost",
		TracerProviderPort:  6831,
		TracerSampleRatio:   1.0,
		MetricProvider:      "stdout",
		MetricProviderHost:  "localhost",
		MetricProviderPort:  9090,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize monitoring: %w", err)
	}

	// Connect to database
	dbPool, err := NewTestPostgreConnection(t, ctx, cfg)
	if err != nil {
		t.Fatal(err)
		return nil, err
	}

	// Connect to Redis
	redisClient, err := NewTestRedisConnection(t, ctx, cfg)
	if err != nil {
		t.Fatal(err)
		return nil, err
	}

	// Initialize repository provider
	repoProvider := repository.NewRepositoryProvider(dbPool)

	// Initialize unit of work
	uow := repository.NewUnitOfWork(dbPool)

	// Initialize resolver provider
	resolverProvider := resolver.NewResolverProvider(
		dbPool,
		redisClient,
		"test:service-access",
		15*time.Minute,
		monitoring.Logger,
		monitoring.Tracer,
	)

	// Initialize UID generator
	uidGenerator := adaptersecurity.NewUIDGenerator()

	// Create event publishers (no-op for tests)
	eventPublisher := publisher.NewNoOpPublisher()

	// Initialize observers (no-op for tests)
	permissionObserver := adapterobserver.NewNoopObserver[signal.SignalPermission]()
	groupObserver := adapterobserver.NewNoopObserver[signal.SignalGroup]()
	groupPermissionObserver := adapterobserver.NewNoopObserver[signal.SignalGroupPermission]()
	roleObserver := adapterobserver.NewNoopObserver[signal.SignalRole]()
	rolePermissionObserver := adapterobserver.NewNoopObserver[signal.SignalRolePermission]()
	subjectObserver := adapterobserver.NewNoopObserver[signal.SignalSubject]()
	accessObserver := adapterobserver.NewNoopObserver[signal.SignalAccessCheck]()

	// Initialize services
	permissionService := service.NewPermissionService(
		uow,
		repoProvider,
		eventPublisher,
		uidGenerator,
		resolverProvider,
		permissionObserver,
	)

	groupService := service.NewGroupService(
		uow,
		repoProvider,
		eventPublisher,
		uidGenerator,
		resolverProvider,
		groupObserver,
		groupPermissionObserver,
	)

	roleService := service.NewRoleService(
		uow,
		repoProvider,
		eventPublisher,
		uidGenerator,
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

	return &TestServices{
		Cfg:               cfg,
		DBPool:            dbPool,
		Redis:             redisClient,
		Monitoring:        monitoring,
		UIDGenerator:      uidGenerator,
		PermissionService: permissionService,
		GroupService:      groupService,
		RoleService:       roleService,
		SubjectService:    subjectService,
		AccessService:     accessService,
	}, nil
}
