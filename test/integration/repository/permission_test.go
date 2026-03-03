package repository

import (
	"context"
	"testing"

	"github.com/adityakw90/service-access/internal/adapter/repository"
	"github.com/adityakw90/service-access/internal/core/domain/model"
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
				UID:         "test-perm-uid-001",
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
			errMsg:  "invalid entity",
		},
		{
			name: "Duplicate Resource and Action",
			input: model.Permission{
				UID:         "test-perm-uid-002",
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
					UID:         "test-perm-uid-original",
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
			uid:     "non-existent-uid",
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
