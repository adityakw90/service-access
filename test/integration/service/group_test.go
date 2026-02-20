package integration

import (
	"context"
	"testing"

	"github.com/adityakw90/service-access/internal/adapter/publisher"
	"github.com/adityakw90/service-access/internal/adapter/repository"
	"github.com/adityakw90/service-access/internal/adapter/security"
	"github.com/adityakw90/service-access/internal/core/domain/param"
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

	uow := repository.NewUnitOfWork(db)
	repos := repository.NewRepositoryProvider(db)

	// Create a mock publisher for testing
	noopPublisher := publisher.NewNoOpPublisher()
	uidGenerator := security.NewUIDGenerator()
	groupService := service.NewGroupService(uow, repos, noopPublisher, uidGenerator)

	// Test: create group
	group, err := groupService.Create(ctx, param.GroupCreateParam{
		Name:        "test-group",
		Description: "Test group for integration",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, group.UID)
	assert.Equal(t, "test-group", group.Name)
}
