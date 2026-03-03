package repository

import (
	"context"
	"testing"

	"github.com/adityakw90/service-access/internal/adapter/repository"
	"github.com/adityakw90/service-access/internal/core/domain/model"
	testutil "github.com/adityakw90/service-access/test/util"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// setupIntegrationTest creates a PostgreSQL connection and truncates all tables.
// Returns a database connection pool that will be cleaned up automatically.
func setupIntegrationTest(t *testing.T) *repository.dbExecutor {
	t.Helper()

	ctx := context.Background()
	cfg, err := testutil.LoadTestConfig(t)
	require.NoError(t, err, "failed to load test config")

	db, err := testutil.NewTestPostgreConnection(t, ctx, cfg)
	require.NoError(t, err, "failed to connect to test database")

	return db
}

// createTestPermission creates a test permission with the given parameters.
// Returns the created permission with ID, UID, and timestamps populated.
func createTestPermission(t *testing.T, db interface{}, resource, action, description string) *model.Permission {
	t.Helper()

	ctx := context.Background()
	repo := repository.NewPermissionRepository(db)

	permission := &model.Permission{
		UID:         uuid.New().String(),
		Resource:    resource,
		Action:      action,
		Description: description,
	}

	err := repo.Create(ctx, permission)
	require.NoError(t, err, "failed to create test permission")

	return permission
}

// createTestGroup creates a test group with the given parameters.
// Returns the created group with ID, UID, and timestamps populated.
func createTestGroup(t *testing.T, db interface{}, name, description string) *model.Group {
	t.Helper()

	ctx := context.Background()
	repo := repository.NewGroupRepository(db)

	group := &model.Group{
		UID:         uuid.New().String(),
		Name:        name,
		Description: description,
	}

	err := repo.Create(ctx, group)
	require.NoError(t, err, "failed to create test group")

	return group
}

// createTestRole creates a test role with the given parameters.
// Returns the created role with ID, UID, and timestamps populated.
func createTestRole(t *testing.T, db interface{}, groupID int64, name, description string) *model.Role {
	t.Helper()

	ctx := context.Background()
	repo := repository.NewRoleRepository(db)

	role := &model.Role{
		UID:         uuid.New().String(),
		GroupID:     groupID,
		Name:        name,
		Description: description,
	}

	err := repo.Create(ctx, role)
	require.NoError(t, err, "failed to create test role")

	return role
}

// createTestSubjectRole creates a test subject role assignment.
// Returns the created subject role with AssignedAt populated.
func createTestSubjectRole(t *testing.T, db interface{}, subjectID, subjectType string, roleID int64) *model.SubjectRole {
	t.Helper()

	ctx := context.Background()
	repo := repository.NewSubjectRepository(db)

	subjectRole := &model.SubjectRole{
		SubjectID:   subjectID,
		SubjectType: subjectType,
		RoleID:      roleID,
	}

	err := repo.Create(ctx, subjectRole)
	require.NoError(t, err, "failed to create test subject role")

	return subjectRole
}
