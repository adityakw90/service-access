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

func TestPermissionRepository_Create(t *testing.T) {
	tests := []struct {
		name    string
		input   model.Permission
		wantErr bool
		errMsg  string
	}{
		{
			name: "Happy Path",
			input: model.Permission{
				UID:         "00000000-0000-0000-0000-000000000001",
				Resource:    "article",
				Action:      "read",
				Description: "Can read articles",
			},
			wantErr: false,
		},
		{
			name: "Empty UID",
			input: model.Permission{
				Resource:    "article",
				Action:      "read",
				Description: "Can read articles",
			},
			wantErr: true,
			errMsg:  "missing required fields",
		},
		{
			name: "Duplicate Resource and Action",
			input: model.Permission{
				UID:         "00000000-0000-0000-0000-000000000002",
				Resource:    "article", // Will be created first
				Action:      "read",
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
			repo := repository.NewPermissionRepository(db)

			// For duplicate test, create the original first
			if tt.name == "Duplicate Resource and Action" {
				original := &model.Permission{
					UID:         "00000000-0000-0000-0000-000000000000",
					Resource:    "article",
					Action:      "read",
					Description: "Original permission",
				}
				err := repo.Create(ctx, original)
				require.NoError(t, err, "failed to create original permission")
			}

			err := repo.Create(ctx, &tt.input)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
				assert.NotZero(t, tt.input.ID)
				assert.NotZero(t, tt.input.CreatedAt)
				assert.NotZero(t, tt.input.UpdatedAt)
			}
		})
	}
}

func TestPermissionRepository_GetByID(t *testing.T) {
	db := setupIntegrationTest(t)
	ctx := context.Background()
	repo := repository.NewPermissionRepository(db)

	// Create a test permission
	perm := createTestPermission(t, db, "document", "write", "Can write documents")

	tests := []struct {
		name    string
		id      int64
		wantErr bool
	}{
		{
			name:    "Found",
			id:      perm.ID,
			wantErr: false,
		},
		{
			name:    "Not Found",
			id:      99999,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetByID(ctx, tt.id)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), "not found")
			} else {
				require.NoError(t, err)
				assert.Equal(t, perm.ID, result.ID)
				assert.Equal(t, perm.UID, result.UID)
				assert.Equal(t, perm.Resource, result.Resource)
				assert.Equal(t, perm.Action, result.Action)
			}
		})
	}
}

func TestPermissionRepository_GetByUID(t *testing.T) {
	db := setupIntegrationTest(t)
	ctx := context.Background()
	repo := repository.NewPermissionRepository(db)

	// Create a test permission
	perm := createTestPermission(t, db, "video", "delete", "Can delete videos")

	tests := []struct {
		name    string
		uid     string
		wantErr bool
	}{
		{
			name:    "Found",
			uid:     perm.UID,
			wantErr: false,
		},
		{
			name:    "Not Found",
			uid:     "00000000-0000-0000-0000-999999999999",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetByUID(ctx, tt.uid)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), "not found")
			} else {
				require.NoError(t, err)
				assert.Equal(t, perm.ID, result.ID)
				assert.Equal(t, perm.UID, result.UID)
			}
		})
	}
}

func TestPermissionRepository_Update(t *testing.T) {
	db := setupIntegrationTest(t)
	ctx := context.Background()
	repo := repository.NewPermissionRepository(db)

	// Create test permission
	perm := createTestPermission(t, db, "image", "view", "Can view images")

	tests := []struct {
		name    string
		setup   func() *model.Permission
		wantErr bool
		errMsg  string
	}{
		{
			name: "Happy Path",
			setup: func() *model.Permission {
				perm.Resource = "photo"
				perm.Action = "download"
				perm.Description = "Can download photos"
				return perm
			},
			wantErr: false,
		},
		{
			name: "Not Found",
			setup: func() *model.Permission {
				return &model.Permission{
					ID:          99999,
					UID:         "00000000-0000-0000-0000-999999999999",
					Resource:    "x",
					Action:      "y",
					Description: "z",
				}
			},
			wantErr: false, // Update doesn't check existence, just returns no error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := tt.setup()
			err := repo.Update(ctx, input)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPermissionRepository_Delete(t *testing.T) {
	db := setupIntegrationTest(t)
	ctx := context.Background()
	repo := repository.NewPermissionRepository(db)

	// Create test permission
	perm := createTestPermission(t, db, "audio", "play", "Can play audio")

	tests := []struct {
		name    string
		id      int64
		wantErr bool
	}{
		{
			name:    "Happy Path",
			id:      perm.ID,
			wantErr: false,
		},
		{
			name:    "Not Found",
			id:      99999,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Delete(ctx, tt.id)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "not found")
			} else {
				require.NoError(t, err)

				// Verify deletion
				_, err := repo.GetByID(ctx, tt.id)
				assert.Error(t, err)
			}
		})
	}
}

func TestPermissionRepository_List(t *testing.T) {
	db := setupIntegrationTest(t)
	ctx := context.Background()
	repo := repository.NewPermissionRepository(db)

	// Create multiple test permissions
	perm1 := createTestPermission(t, db, "book", "read", "Read books")
	perm2 := createTestPermission(t, db, "book", "write", "Write books")
	perm3 := createTestPermission(t, db, "magazine", "read", "Read magazines")

	tests := []struct {
		name         string
		pagination   *param.PaginationParam
		filter       *param.PermissionListFilterParam
		minTotal     int64
		shouldFind   []string // UIDs that should be in results
		shouldNotFind []string // UIDs that should NOT be in results
	}{
		{
			name:     "List All",
			filter:   &param.PermissionListFilterParam{},
			minTotal: 3,
		},
		{
			name: "Filter By Resource",
			filter: &param.PermissionListFilterParam{
				Resource: strPtr("book"),
			},
			minTotal:     2,
			shouldFind:   []string{perm1.UID, perm2.UID},
			shouldNotFind: []string{perm3.UID},
		},
		{
			name: "Filter By Action",
			filter: &param.PermissionListFilterParam{
				Action: strPtr("read"),
			},
			minTotal:     2,
			shouldFind:   []string{perm1.UID, perm3.UID},
			shouldNotFind: []string{perm2.UID},
		},
		{
			name: "Filter By IDs",
			filter: &param.PermissionListFilterParam{
				IDs: []int64{perm1.ID, perm2.ID},
			},
			minTotal:     2,
			shouldFind:   []string{perm1.UID, perm2.UID},
			shouldNotFind: []string{perm3.UID},
		},
		{
			name: "Filter By Query",
			filter: &param.PermissionListFilterParam{
				Query: strPtr("book"),
			},
			minTotal:     2,
			shouldFind:   []string{perm1.UID, perm2.UID},
			shouldNotFind: []string{perm3.UID},
		},
		{
			name: "Pagination",
			filter: &param.PermissionListFilterParam{},
			pagination: &param.PaginationParam{
				Limit: intPtr(2),
				Page:  intPtr(1),
			},
			minTotal: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.List(ctx, tt.pagination, tt.filter)

			require.NoError(t, err)
			assert.GreaterOrEqual(t, result.Meta.Total, tt.minTotal)

			// Check specific UIDs
			resultUIDs := make(map[string]bool)
			for _, item := range result.Items {
				resultUIDs[item.UID] = true
			}

			for _, uid := range tt.shouldFind {
				assert.True(t, resultUIDs[uid], "UID %s should be in results", uid)
			}
			for _, uid := range tt.shouldNotFind {
				assert.False(t, resultUIDs[uid], "UID %s should NOT be in results", uid)
			}
		})
	}
}

// Helper functions
func strPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func int64Ptr(i int64) *int64 {
	return &i
}
