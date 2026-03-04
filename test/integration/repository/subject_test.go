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

func TestSubjectRepository_Create(t *testing.T) {
	db := setupIntegrationTest(t)
	ctx := context.Background()
	repo := repository.NewSubjectRepository(db)

	group := createTestGroup(t, db, "subject-group", "Subject test group")
	role := createTestRole(t, db, group.ID, "subject-role", "Subject test role")

	tests := []struct {
		name         string
		input        model.SubjectRole
		wantErr      bool
		errMsg       string
	}{
		{
			name: "Happy Path",
			input: model.SubjectRole{
				SubjectID:   "user-123",
				SubjectType: "user",
				RoleID:      role.ID,
			},
			wantErr: false,
		},
		{
			name: "Non-existent Role",
			input: model.SubjectRole{
				SubjectID:   "user-456",
				SubjectType: "user",
				RoleID:      99999,
			},
			wantErr: true,
			errMsg:  "role with id",
		},
		{
			name: "Duplicate Assignment",
			input: model.SubjectRole{
				SubjectID:   "user-789",
				SubjectType: "user",
				RoleID:      role.ID,
			},
			wantErr: true,
			errMsg:  "duplicate key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For duplicate test, create original first
			if tt.name == "Duplicate Assignment" {
				original := &model.SubjectRole{
					SubjectID:   "user-789",
					SubjectType: "user",
					RoleID:      role.ID,
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
				assert.NotZero(t, tt.input.AssignedAt)
			}
		})
	}
}

func TestSubjectRepository_GetRoles(t *testing.T) {
	db := setupIntegrationTest(t)
	ctx := context.Background()
	repo := repository.NewSubjectRepository(db)

	group := createTestGroup(t, db, "get-roles-group", "Get roles group")
	role1 := createTestRole(t, db, group.ID, "get-roles-role-1", "Get roles role 1")
	role2 := createTestRole(t, db, group.ID, "get-roles-role-2", "Get roles role 2")

	subjectID := "user-get-roles"
	subjectType := "user"

	// Assign multiple roles to subject
	err := repo.Create(ctx, &model.SubjectRole{
		SubjectID:   subjectID,
		SubjectType: subjectType,
		RoleID:      role1.ID,
	})
	require.NoError(t, err)

	err = repo.Create(ctx, &model.SubjectRole{
		SubjectID:   subjectID,
		SubjectType: subjectType,
		RoleID:      role2.ID,
	})
	require.NoError(t, err)

	// Get roles
	roles, err := repo.GetRoles(ctx, subjectID, subjectType)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(roles), 2)
}

func TestSubjectRepository_Delete(t *testing.T) {
	db := setupIntegrationTest(t)
	ctx := context.Background()
	repo := repository.NewSubjectRepository(db)

	group := createTestGroup(t, db, "delete-subject-group", "Delete subject group")
	role := createTestRole(t, db, group.ID, "delete-subject-role", "Delete subject role")

	subjectID := "user-delete"
	subjectType := "user"

	// Create assignment
	err := repo.Create(ctx, &model.SubjectRole{
		SubjectID:   subjectID,
		SubjectType: subjectType,
		RoleID:      role.ID,
	})
	require.NoError(t, err)

	// Delete assignment
	err = repo.Delete(ctx, subjectID, subjectType, role.ID)
	require.NoError(t, err)

	// Verify deletion
	roles, err := repo.GetRoles(ctx, subjectID, subjectType)
	require.NoError(t, err)
	assert.Equal(t, 0, len(roles))
}

func TestSubjectRepository_List(t *testing.T) {
	db := setupIntegrationTest(t)
	ctx := context.Background()
	repo := repository.NewSubjectRepository(db)

	group := createTestGroup(t, db, "list-subject-group", "List subject group")
	role1 := createTestRole(t, db, group.ID, "list-subject-role-1", "List subject role 1")
	role2 := createTestRole(t, db, group.ID, "list-subject-role-2", "List subject role 2")

	// Create multiple subject role assignments
	err := repo.Create(ctx, &model.SubjectRole{
		SubjectID:   "user-list-1",
		SubjectType: "user",
		RoleID:      role1.ID,
	})
	require.NoError(t, err)

	err = repo.Create(ctx, &model.SubjectRole{
		SubjectID:   "user-list-2",
		SubjectType: "user",
		RoleID:      role2.ID,
	})
	require.NoError(t, err)

	err = repo.Create(ctx, &model.SubjectRole{
		SubjectID:   "service-list-1",
		SubjectType: "service",
		RoleID:      role1.ID,
	})
	require.NoError(t, err)

	tests := []struct {
		name       string
		filter     *param.SubjectListFilterParam
		minCount   int
		shouldFind []string // Subject IDs that should be in results
	}{
		{
			name:     "List All",
			filter:   &param.SubjectListFilterParam{},
			minCount: 3,
		},
		{
			name: "Filter By SubjectID",
			filter: &param.SubjectListFilterParam{
				SubjectID: strPtr("user-list-1"),
			},
			minCount:   1,
			shouldFind: []string{"user-list-1"},
		},
		{
			name: "Filter By SubjectType",
			filter: &param.SubjectListFilterParam{
				SubjectType: strPtr("user"),
			},
			minCount: 2,
		},
		{
			name: "Filter By RoleID",
			filter: &param.SubjectListFilterParam{
				RoleID: int64Ptr(role1.ID),
			},
			minCount: 2,
		},
		{
			name: "Filter By RoleUID",
			filter: &param.SubjectListFilterParam{
				RoleUID: strPtr(role2.UID),
			},
			minCount: 1,
		},
		{
			name: "Filter By Query",
			filter: &param.SubjectListFilterParam{
				Query: strPtr("user-list"),
			},
			minCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.List(ctx, nil, tt.filter)
			require.NoError(t, err)
			assert.GreaterOrEqual(t, len(result.Items), tt.minCount)

			if len(tt.shouldFind) > 0 {
				resultSubjectIDs := make(map[string]bool)
				for _, item := range result.Items {
					resultSubjectIDs[item.SubjectID] = true
				}
				for _, subjectID := range tt.shouldFind {
					assert.True(t, resultSubjectIDs[subjectID], "SubjectID %s should be in results", subjectID)
				}
			}
		})
	}
}
