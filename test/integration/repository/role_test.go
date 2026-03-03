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
