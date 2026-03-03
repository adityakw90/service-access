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

func TestRoleRepository_Create(t *testing.T) {
	db := setupIntegrationTest(t)
	ctx := context.Background()
	repo := repository.NewRoleRepository(db)

	group := createTestGroup(t, db, "admin-group", "Admin group")

	tests := []struct {
		name    string
		input   model.Role
		wantErr bool
		errMsg  string
	}{
		{
			name: "Happy Path",
			input: model.Role{
				UID:         "test-role-uid-001",
				GroupID:     group.ID,
				Name:        "super-admin",
				Description: "Super admin",
			},
			wantErr: false,
		},
		{
			name: "Empty UID",
			input: model.Role{
				GroupID:     group.ID,
				Name:        "admin",
				Description: "Admin",
			},
			wantErr: true,
			errMsg:  "invalid entity",
		},
		{
			name: "Non-existent Group",
			input: model.Role{
				UID:         "test-role-uid-002",
				GroupID:     99999,
				Name:        "moderator",
				Description: "Moderator",
			},
			wantErr: true,
			errMsg:  "group with id",
		},
		{
			name: "Duplicate Name in Group",
			input: model.Role{
				UID:         "test-role-uid-003",
				GroupID:     group.ID,
				Name:        "super-admin", // Duplicate
				Description: "Duplicate role",
			},
			wantErr: true,
			errMsg:  "already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For duplicate test, create the original first
			if tt.name == "Duplicate Name in Group" {
				original := &model.Role{
					UID:         "test-role-uid-original",
					GroupID:     group.ID,
					Name:        "super-admin",
					Description: "Original role",
				}
				err := repo.Create(ctx, original)
				require.NoError(t, err)
			}

			err := repo.Create(ctx, &tt.input)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
				assert.NotZero(t, tt.input.ID)
			}
		})
	}
}

func TestRoleRepository_GetByID(t *testing.T) {
	db := setupIntegrationTest(t)
	ctx := context.Background()
	repo := repository.NewRoleRepository(db)

	group := createTestGroup(t, db, "editor-group", "Editors")
	role := createTestRole(t, db, group.ID, "senior-editor", "Senior editor")

	result, err := repo.GetByID(ctx, role.ID)
	require.NoError(t, err)
	assert.Equal(t, role.ID, result.ID)
	assert.Equal(t, role.Name, result.Name)
	assert.Equal(t, role.GroupID, result.GroupID)
}

func TestRoleRepository_GetByUID(t *testing.T) {
	db := setupIntegrationTest(t)
	ctx := context.Background()
	repo := repository.NewRoleRepository(db)

	group := createTestGroup(t, db, "writer-group", "Writers")
	role := createTestRole(t, db, group.ID, "content-writer", "Content writer")

	result, err := repo.GetByUID(ctx, role.UID)
	require.NoError(t, err)
	assert.Equal(t, role.UID, result.UID)
	assert.Equal(t, role.Name, result.Name)
}

func TestRoleRepository_Update(t *testing.T) {
	db := setupIntegrationTest(t)
	ctx := context.Background()
	repo := repository.NewRoleRepository(db)

	group := createTestGroup(t, db, "updater-group", "Updaters")
	role := createTestRole(t, db, group.ID, "junior-writer", "Junior writer")

	role.Name = "lead-writer"
	role.Description = "Lead writer"

	err := repo.Update(ctx, role)
	require.NoError(t, err)

	updated, err := repo.GetByID(ctx, role.ID)
	require.NoError(t, err)
	assert.Equal(t, "lead-writer", updated.Name)
}

func TestRoleRepository_Delete(t *testing.T) {
	db := setupIntegrationTest(t)
	ctx := context.Background()
	repo := repository.NewRoleRepository(db)

	group := createTestGroup(t, db, "deleter-group", "Deleters")
	role := createTestRole(t, db, group.ID, "temp-role", "Temporary role")

	err := repo.Delete(ctx, role.ID)
	require.NoError(t, err)

	_, err = repo.GetByID(ctx, role.ID)
	assert.Error(t, err)
}

func TestRoleRepository_List(t *testing.T) {
	db := setupIntegrationTest(t)
	ctx := context.Background()
	repo := repository.NewRoleRepository(db)

	group1 := createTestGroup(t, db, "group-alpha", "Alpha")

	role1 := createTestRole(t, db, group1.ID, "role-alpha-1", "Alpha role 1")
	role2 := createTestRole(t, db, group1.ID, "role-alpha-2", "Alpha role 2")

	tests := []struct {
		name       string
		filter     *param.RoleListFilterParam
		minCount   int
		shouldFind []string
	}{
		{
			name:     "List All",
			filter:   &param.RoleListFilterParam{},
			minCount: 3,
		},
		{
			name: "Filter By GroupID",
			filter: &param.RoleListFilterParam{
				GroupID: int64Ptr(group1.ID),
			},
			minCount:   2,
			shouldFind: []string{role1.UID, role2.UID},
		},
		{
			name: "Filter By Name",
			filter: &param.RoleListFilterParam{
				Name: strPtr("role-alpha-1"),
			},
			minCount:   1,
			shouldFind: []string{role1.UID},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.List(ctx, nil, tt.filter)
			require.NoError(t, err)
			assert.GreaterOrEqual(t, len(result.Items), tt.minCount)

			if len(tt.shouldFind) > 0 {
				resultUIDs := make(map[string]bool)
				for _, item := range result.Items {
					resultUIDs[item.UID] = true
				}
				for _, uid := range tt.shouldFind {
					assert.True(t, resultUIDs[uid], "UID %s should be in results", uid)
				}
			}
		})
	}
}

func TestRoleRepository_AddPermission(t *testing.T) {
	db := setupIntegrationTest(t)
	ctx := context.Background()
	repo := repository.NewRoleRepository(db)
	groupRepo := repository.NewGroupRepository(db)

	group := createTestGroup(t, db, "perm-test-group", "Permission test group")
	role := createTestRole(t, db, group.ID, "perm-test-role", "Permission test role")
	perm := createTestPermission(t, db, "test-resource", "test-action", "Test permission")

	// Add permission to group first
	groupPermUID := "test-group-perm-uid"
	err := groupRepo.AddPermission(ctx, group.ID, perm.ID, groupPermUID)
	require.NoError(t, err)

	// Get the group permission ID
	groupPerms, err := groupRepo.ListPermission(ctx, group.ID, nil, &param.GroupPermissionListFilterParam{})
	require.NoError(t, err)
	require.Greater(t, len(groupPerms.Items), 0)
	groupPermID := groupPerms.Items[0].ID

	tests := []struct {
		name              string
		roleID            int64
		groupPermissionID int64
		wantErr           bool
		errMsg            string
	}{
		{
			name:              "Happy Path",
			roleID:            role.ID,
			groupPermissionID: groupPermID,
			wantErr:           false,
		},
		{
			name:              "Non-existent Role",
			roleID:            99999,
			groupPermissionID: groupPermID,
			wantErr:           true,
			errMsg:            "role with id",
		},
		{
			name:              "Non-existent Group Permission",
			roleID:            role.ID,
			groupPermissionID: 99999,
			wantErr:           true,
			errMsg:            "group permission with id",
		},
		{
			name:              "Duplicate (ON CONFLICT DO NOTHING)",
			roleID:            role.ID,
			groupPermissionID: groupPermID,
			wantErr:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.AddPermission(ctx, tt.roleID, tt.groupPermissionID)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRoleRepository_RemovePermission(t *testing.T) {
	db := setupIntegrationTest(t)
	ctx := context.Background()
	repo := repository.NewRoleRepository(db)
	groupRepo := repository.NewGroupRepository(db)

	group := createTestGroup(t, db, "remove-test-group", "Remove test group")
	role := createTestRole(t, db, group.ID, "remove-test-role", "Remove test role")
	perm := createTestPermission(t, db, "remove-resource", "remove-action", "Remove test permission")

	// Add permission to group and role
	groupPermUID := "test-remove-group-perm-uid"
	err := groupRepo.AddPermission(ctx, group.ID, perm.ID, groupPermUID)
	require.NoError(t, err)

	groupPerms, err := groupRepo.ListPermission(ctx, group.ID, nil, &param.GroupPermissionListFilterParam{})
	require.NoError(t, err)
	groupPermID := groupPerms.Items[0].ID

	err = repo.AddPermission(ctx, role.ID, groupPermID)
	require.NoError(t, err)

	// Test removal
	err = repo.RemovePermission(ctx, role.ID, groupPermID)
	require.NoError(t, err)

	// Verify removal (should error when trying to remove again)
	err = repo.RemovePermission(ctx, role.ID, groupPermID)
	assert.Error(t, err)
}

func TestRoleRepository_ListPermission(t *testing.T) {
	db := setupIntegrationTest(t)
	ctx := context.Background()
	repo := repository.NewRoleRepository(db)
	groupRepo := repository.NewGroupRepository(db)

	group := createTestGroup(t, db, "list-perm-group", "List permission group")
	role := createTestRole(t, db, group.ID, "list-perm-role", "List permission role")
	perm1 := createTestPermission(t, db, "list-resource-1", "list-action-1", "List test permission 1")
	perm2 := createTestPermission(t, db, "list-resource-2", "list-action-2", "List test permission 2")

	// Add permissions to group
	err := groupRepo.AddPermission(ctx, group.ID, perm1.ID, "test-list-group-perm-uid-1")
	require.NoError(t, err)
	err = groupRepo.AddPermission(ctx, group.ID, perm2.ID, "test-list-group-perm-uid-2")
	require.NoError(t, err)

	// Get group permission IDs
	groupPerms, err := groupRepo.ListPermission(ctx, group.ID, nil, &param.GroupPermissionListFilterParam{})
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(groupPerms.Items), 2)

	// Add permissions to role
	for _, gp := range groupPerms.Items {
		err = repo.AddPermission(ctx, role.ID, gp.ID)
		require.NoError(t, err)
	}

	// List role permissions
	result, err := repo.ListPermission(ctx, role.ID, nil, &param.RolePermissionListFilterParam{})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(result.Items), 2)
}
