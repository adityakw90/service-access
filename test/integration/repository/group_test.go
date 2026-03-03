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

func TestGroupRepository_Create(t *testing.T) {
	tests := []struct {
		name    string
		input   model.Group
		wantErr bool
		errMsg  string
	}{
		{
			name: "Happy Path",
			input: model.Group{
				UID:         "00000000-0000-0000-0000-000000000001",
				Name:        "editors",
				Description: "Can edit content",
			},
			wantErr: false,
		},
		{
			name: "Empty UID",
			input: model.Group{
				Name:        "editors",
				Description: "Can edit content",
			},
			wantErr: true,
			errMsg:  "missing required fields",
		},
		{
			name: "Duplicate Name",
			input: model.Group{
				UID:         "00000000-0000-0000-0000-000000000002",
				Name:        "admins",
				Description: "Duplicate",
			},
			wantErr: true,
			errMsg:  "already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupIntegrationTest(t)
			ctx := context.Background()
			repo := repository.NewGroupRepository(db)

			// For duplicate test, create the original first
			if tt.name == "Duplicate Name" {
				original := &model.Group{
					UID:         "00000000-0000-0000-0000-000000000000",
					Name:        "admins",
					Description: "Original group",
				}
				err := repo.Create(ctx, original)
				require.NoError(t, err, "failed to create original group")
			}

			err := repo.Create(ctx, &tt.input)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
				assert.NotZero(t, tt.input.ID)
				assert.NotZero(t, tt.input.CreatedAt)
			}
		})
	}
}

func TestGroupRepository_GetByID(t *testing.T) {
	db := setupIntegrationTest(t)
	ctx := context.Background()
	repo := repository.NewGroupRepository(db)

	group := createTestGroup(t, db, "moderators", "Can moderate content")

	tests := []struct {
		name    string
		id      int64
		wantErr bool
	}{
		{"Found", group.ID, false},
		{"Not Found", 99999, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetByID(ctx, tt.id)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "not found")
			} else {
				require.NoError(t, err)
				assert.Equal(t, group.ID, result.ID)
				assert.Equal(t, group.Name, result.Name)
			}
		})
	}
}

func TestGroupRepository_GetByUID(t *testing.T) {
	db := setupIntegrationTest(t)
	ctx := context.Background()
	repo := repository.NewGroupRepository(db)

	group := createTestGroup(t, db, "contributors", "Can contribute content")

	result, err := repo.GetByUID(ctx, group.UID)
	require.NoError(t, err)
	assert.Equal(t, group.UID, result.UID)
	assert.Equal(t, group.Name, result.Name)
}

func TestGroupRepository_Update(t *testing.T) {
	db := setupIntegrationTest(t)
	ctx := context.Background()
	repo := repository.NewGroupRepository(db)

	group := createTestGroup(t, db, "subscribers", "Can subscribe")

	group.Name = "premium-subscribers"
	group.Description = "Premium subscribers"

	err := repo.Update(ctx, group)
	require.NoError(t, err)

	// Verify update
	updated, err := repo.GetByID(ctx, group.ID)
	require.NoError(t, err)
	assert.Equal(t, "premium-subscribers", updated.Name)
}

func TestGroupRepository_Delete(t *testing.T) {
	db := setupIntegrationTest(t)
	ctx := context.Background()
	repo := repository.NewGroupRepository(db)

	group := createTestGroup(t, db, "guests", "Guest users")

	err := repo.Delete(ctx, group.ID)
	require.NoError(t, err)

	// Verify deletion
	_, err = repo.GetByID(ctx, group.ID)
	assert.Error(t, err)
}

func TestGroupRepository_List(t *testing.T) {
	db := setupIntegrationTest(t)
	ctx := context.Background()
	repo := repository.NewGroupRepository(db)

	group1 := createTestGroup(t, db, "alpha", "Alpha group")
	group2 := createTestGroup(t, db, "beta", "Beta group")
	group3 := createTestGroup(t, db, "gamma", "Gamma group")

	tests := []struct {
		name       string
		filter     *param.GroupListFilterParam
		shouldFind []string
	}{
		{
			name:       "List All",
			filter:     &param.GroupListFilterParam{},
			shouldFind: []string{group1.UID, group2.UID, group3.UID},
		},
		{
			name: "Filter By Name",
			filter: &param.GroupListFilterParam{
				Name: strPtr("alpha"),
			},
			shouldFind: []string{group1.UID},
		},
		{
			name: "Filter By IDs",
			filter: &param.GroupListFilterParam{
				IDs: []int64{group1.ID, group2.ID},
			},
			shouldFind: []string{group1.UID, group2.UID},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.List(ctx, nil, tt.filter)
			require.NoError(t, err)
			assert.NotZero(t, result.Meta.Total)

			resultUIDs := make(map[string]bool)
			for _, item := range result.Items {
				resultUIDs[item.UID] = true
			}

			for _, uid := range tt.shouldFind {
				assert.True(t, resultUIDs[uid], "UID %s should be in results", uid)
			}
		})
	}
}

func TestGroupRepository_AddPermission(t *testing.T) {
	db := setupIntegrationTest(t)
	ctx := context.Background()
	repo := repository.NewGroupRepository(db)

	group := createTestGroup(t, db, "readers", "Can read")
	perm := createTestPermission(t, db, "blog", "read", "Read blogs")

	uid := "test-group-perm-uid-001"

	tests := []struct {
		name         string
		groupID      int64
		permissionID int64
		uid          string
		wantErr      bool
		errMsg       string
	}{
		{
			name:         "Happy Path",
			groupID:      group.ID,
			permissionID: perm.ID,
			uid:          uid,
			wantErr:      false,
		},
		{
			name:         "Non-existent Group",
			groupID:      99999,
			permissionID: perm.ID,
			uid:          "test-uid-002",
			wantErr:      true,
			errMsg:       "group with id",
		},
		{
			name:         "Non-existent Permission",
			groupID:      group.ID,
			permissionID: 99999,
			uid:          "test-uid-003",
			wantErr:      true,
			errMsg:       "permission with id",
		},
		{
			name:         "Duplicate (ON CONFLICT DO NOTHING)",
			groupID:      group.ID,
			permissionID: perm.ID,
			uid:          "test-uid-004",
			wantErr:      false, // Should not error due to ON CONFLICT
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.AddPermission(ctx, tt.groupID, tt.permissionID, tt.uid)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGroupRepository_RemovePermission(t *testing.T) {
	db := setupIntegrationTest(t)
	ctx := context.Background()
	repo := repository.NewGroupRepository(db)

	group := createTestGroup(t, db, "removers", "Can remove")
	perm := createTestPermission(t, db, "comment", "delete", "Delete comments")

	// Add permission first
	err := repo.AddPermission(ctx, group.ID, perm.ID, "test-remove-uid-001")
	require.NoError(t, err)

	tests := []struct {
		name         string
		groupID      int64
		permissionID int64
		wantErr      bool
	}{
		{
			name:         "Happy Path",
			groupID:      group.ID,
			permissionID: perm.ID,
			wantErr:      false,
		},
		{
			name:         "Non-existent Association",
			groupID:      group.ID,
			permissionID: 99999,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.RemovePermission(ctx, tt.groupID, tt.permissionID)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "not found")
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGroupRepository_ListPermission(t *testing.T) {
	db := setupIntegrationTest(t)
	ctx := context.Background()
	repo := repository.NewGroupRepository(db)

	group := createTestGroup(t, db, "viewers", "Can view")
	perm1 := createTestPermission(t, db, "page", "view", "View pages")
	perm2 := createTestPermission(t, db, "post", "view", "View posts")

	// Add permissions
	err := repo.AddPermission(ctx, group.ID, perm1.ID, "test-list-perm-uid-001")
	require.NoError(t, err)
	err = repo.AddPermission(ctx, group.ID, perm2.ID, "test-list-perm-uid-002")
	require.NoError(t, err)

	tests := []struct {
		name     string
		filter   *param.GroupPermissionListFilterParam
		minCount int
	}{
		{
			name:     "List All",
			filter:   &param.GroupPermissionListFilterParam{},
			minCount: 2,
		},
		{
			name: "Filter By Resource",
			filter: &param.GroupPermissionListFilterParam{
				Resource: strPtr("page"),
			},
			minCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.ListPermission(ctx, group.ID, nil, tt.filter)
			require.NoError(t, err)
			assert.GreaterOrEqual(t, len(result.Items), tt.minCount)
		})
	}
}
