package repository

import (
	"context"
	"testing"

	"github.com/adityakw90/service-access/internal/adapter/repository"
	"github.com/adityakw90/service-access/internal/core/domain/model"
	"github.com/adityakw90/service-access/internal/core/domain/param"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRoleRepository_GetAllPermissions tests fetching all permissions for a role.
// Permissions are fetched through the join: role → role_permission → group_permission → permission.
func TestRoleRepository_GetAllPermissions(t *testing.T) {
	db := setupIntegrationTest(t)
	ctx := context.Background()

	repo := repository.NewRoleRepository(db)
	groupRepo := repository.NewGroupRepository(db)

	// Create test data
	group := createTestGroup(t, db, "test-group", "Test group")
	role := createTestRole(t, db, group.ID, "test-role", "Test role")
	perm1 := createTestPermission(t, db, "doc", "read", "Read documents")
	perm2 := createTestPermission(t, db, "doc", "write", "Write documents")
	perm3 := createTestPermission(t, db, "user", "read", "Read users")

	// Add permissions to group (creates group_permission entries)
	require.NoError(t, groupRepo.AddPermission(ctx, group.ID, perm1.ID, "00000000-0000-0000-0000-000000000011"))
	require.NoError(t, groupRepo.AddPermission(ctx, group.ID, perm2.ID, "00000000-0000-0000-0000-000000000012"))
	require.NoError(t, groupRepo.AddPermission(ctx, group.ID, perm3.ID, "00000000-0000-0000-0000-000000000013"))

	// Get group permission IDs
	groupPerms, err := groupRepo.ListPermission(ctx, group.ID, nil, &param.GroupPermissionListFilterParam{})
	require.NoError(t, err)
	require.Len(t, groupPerms.Items, 3)

	// Assign group permissions to role (creates role_permission entries)
	for _, gp := range groupPerms.Items {
		require.NoError(t, repo.AddPermission(ctx, role.ID, gp.ID))
	}

	// Execute
	perms, err := repo.GetAllPermissions(ctx, role.ID)
	require.NoError(t, err)
	require.Len(t, perms, 3)

	// Verify sorting by UID (must be sorted ascending)
	for i := 1; i < len(perms); i++ {
		assert.True(t, perms[i].UID >= perms[i-1].UID, "permissions not sorted by UID")
	}

	// Verify each permission matches one of the created ones
	byID := make(map[int64]model.Permission)
	for _, p := range perms {
		byID[p.ID] = p
	}

	check := func(p *model.Permission) {
		got, ok := byID[p.ID]
		require.True(t, ok, "permission ID %d not found", p.ID)
		assert.Equal(t, p.UID, got.UID)
		assert.Equal(t, p.Resource, got.Resource)
		assert.Equal(t, p.Action, got.Action)
		assert.Equal(t, p.Description, got.Description)
		assert.NotZero(t, got.CreatedAt)
		assert.NotZero(t, got.UpdatedAt)
	}

	check(perm1)
	check(perm2)
	check(perm3)
}

// TestRoleRepository_GetAllPermissions_Empty tests when the role has no permissions.
func TestRoleRepository_GetAllPermissions_Empty(t *testing.T) {
	db := setupIntegrationTest(t)
	ctx := context.Background()

	repo := repository.NewRoleRepository(db)

	group := createTestGroup(t, db, "empty-group", "Empty group")
	role := createTestRole(t, db, group.ID, "empty-role", "No permissions")

	perms, err := repo.GetAllPermissions(ctx, role.ID)
	require.NoError(t, err)
	require.Empty(t, perms)
}
