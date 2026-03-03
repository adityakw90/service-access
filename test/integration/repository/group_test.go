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
				UID:         "test-group-uid-001",
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
			errMsg:  "invalid entity",
		},
		{
			name: "Duplicate Name",
			input: model.Group{
				UID:         "test-group-uid-002",
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
					UID:         "test-group-uid-original",
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
