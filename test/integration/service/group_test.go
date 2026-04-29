package service

import (
	"context"
	"testing"
	"time"

	"github.com/adityakw90/service-access/internal/adapter/event"
	adapterexecutor "github.com/adityakw90/service-access/internal/adapter/executor"
	"github.com/adityakw90/service-access/internal/adapter/observer"
	"github.com/adityakw90/service-access/internal/adapter/repository"
	adapterResolver "github.com/adityakw90/service-access/internal/adapter/resolver"
	"github.com/adityakw90/service-access/internal/adapter/security"
	"github.com/adityakw90/service-access/internal/infra"
	"github.com/adityakw90/service-access/internal/core/domain/param"
	"github.com/adityakw90/service-access/internal/core/domain/signal"
	"github.com/adityakw90/service-access/internal/core/service"
	testutil "github.com/adityakw90/service-access/test/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_GroupService_Create(t *testing.T) {
	ctx := context.Background()
	cfg, err := testutil.LoadTestConfig(t)
	if err != nil {
		t.Fatalf("failed to load test config: %v", err)
	}
	db, err := testutil.NewTestPostgreConnection(t, ctx, cfg)
	if err != nil {
		t.Fatalf("failed to connect to test postgres: %v", err)
	}
	redisClient, err := testutil.NewTestRedisConnection(t, ctx, cfg)
	if err != nil {
		t.Fatalf("failed to connect to test redis: %v", err)
	}

	uow := repository.NewUnitOfWork(db)
	repos := repository.NewRepositoryProvider(db)

	// Create a mock publisher for testing
	noopPublisher := event.NewNoOpPublisher()
	uidGenerator := security.NewUIDGenerator()

	// Create resolver provider with 5 minute TTL
	resolverProvider := adapterResolver.NewResolverProvider(db, redisClient, "service-access", 5*time.Minute, nil, nil)

	// Create executor (no-op for tests)
	exc := adapterexecutor.NewServiceExecutor(infra.NewNoopLogger(), infra.NewNoopTracer())

	// Create noop observer
	groupObserver := observer.NewNoopObserver[signal.SignalGroup]()
	groupPermissionObserver := observer.NewNoopObserver[signal.SignalGroupPermission]()

	// Create service with all dependencies
	groupService := service.NewGroupService(uow, repos, noopPublisher, uidGenerator, resolverProvider, exc, groupObserver, groupPermissionObserver)

	// Test: create group
	group, err := groupService.Create(ctx, param.GroupCreateParam{
		Name:        "test-group",
		Description: "Test group for integration",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, group.UID)
	assert.Equal(t, "test-group", group.Name)
}
