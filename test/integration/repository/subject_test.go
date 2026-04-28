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
		name    string
		input   model.SubjectRole
		wantErr bool
		errMsg  string
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
		pagination *param.PaginationParam
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
		{
			name: "With Default Pagination (created_at OrderBy should fallback)",
			pagination: &param.PaginationParam{
				Page:    func() *int { i := 1; return &i }(),
				Limit:   func() *int { i := 10; return &i }(),
				Sort:    func() *string { s := "asc"; return &s }(),
				OrderBy: func() *string { s := "created_at"; return &s }(), // Invalid column - should fallback to assigned_at
			},
			filter:   &param.SubjectListFilterParam{},
			minCount: 3,
		},
		{
			name: "With Valid OrderBy (subject_id)",
			pagination: &param.PaginationParam{
				Page:    func() *int { i := 1; return &i }(),
				Limit:   func() *int { i := 10; return &i }(),
				Sort:    func() *string { s := "asc"; return &s }(),
				OrderBy: func() *string { s := "subject_id"; return &s }(), // Valid column
			},
			filter:   &param.SubjectListFilterParam{},
			minCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.List(ctx, tt.pagination, tt.filter)
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

func TestSubjectRepository_GetAllRoles(t *testing.T) {
	db := setupIntegrationTest(t)
	ctx := context.Background()
	repo := repository.NewSubjectRepository(db)

	group := createTestGroup(t, db, "get-all-roles-group", "Get all roles group")
	role1 := createTestRole(t, db, group.ID, "get-all-role-1", "Get all role 1")
	role2 := createTestRole(t, db, group.ID, "get-all-role-2", "Get all role 2")
	role3 := createTestRole(t, db, group.ID, "get-all-role-3", "Get all role 3")

	subjectID := "user-get-all-roles"
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

	err = repo.Create(ctx, &model.SubjectRole{
		SubjectID:   subjectID,
		SubjectType: subjectType,
		RoleID:      role3.ID,
	})
	require.NoError(t, err)

	// GetAllRoles should return distinct Role objects (not SubjectRole)
	roles, err := repo.GetAllRoles(ctx, subjectID, subjectType)
	require.NoError(t, err)
	assert.Equal(t, 3, len(roles))

	// Verify Role structure (not SubjectRole)
	for _, role := range roles {
		assert.NotZero(t, role.ID)
		assert.NotZero(t, role.UID)
		assert.NotZero(t, role.GroupID)
		assert.NotZero(t, role.GroupUID)
		assert.NotZero(t, role.Name)
		assert.NotZero(t, role.Description)
		assert.NotZero(t, role.CreatedAt)
		assert.NotZero(t, role.UpdatedAt)
	}

	// Verify order by UID
	for i := 1; i < len(roles); i++ {
		assert.True(t, roles[i].UID >= roles[i-1].UID, "Roles should be sorted by UID")
	}
}

func TestSubjectRepository_GetAllRoles_Empty(t *testing.T) {
	db := setupIntegrationTest(t)
	ctx := context.Background()
	repo := repository.NewSubjectRepository(db)

	subjectID := "user-nonexistent"
	subjectType := "user"

	roles, err := repo.GetAllRoles(ctx, subjectID, subjectType)
	require.NoError(t, err)
	assert.Equal(t, 0, len(roles))
}

func TestSubjectRepository_GetAllGroups(t *testing.T) {
	db := setupIntegrationTest(t)
	ctx := context.Background()
	repo := repository.NewSubjectRepository(db)

	// Create multiple groups and roles
	group1 := createTestGroup(t, db, "get-all-groups-group-1", "Get all groups group 1")
	group2 := createTestGroup(t, db, "get-all-groups-group-2", "Get all groups group 2")
	group3 := createTestGroup(t, db, "get-all-groups-group-3", "Get all groups group 3")

	role1 := createTestRole(t, db, group1.ID, "get-all-group-role-1", "Get all group role 1")
	role2 := createTestRole(t, db, group1.ID, "get-all-group-role-2", "Get all group role 2") // Same group as role1
	role3 := createTestRole(t, db, group2.ID, "get-all-group-role-3", "Get all group role 3")
	role4 := createTestRole(t, db, group3.ID, "get-all-group-role-4", "Get all group role 4")

	subjectID := "user-get-all-groups"
	subjectType := "user"

	// Assign roles from multiple groups, including multiple roles from same group
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

	err = repo.Create(ctx, &model.SubjectRole{
		SubjectID:   subjectID,
		SubjectType: subjectType,
		RoleID:      role3.ID,
	})
	require.NoError(t, err)

	err = repo.Create(ctx, &model.SubjectRole{
		SubjectID:   subjectID,
		SubjectType: subjectType,
		RoleID:      role4.ID,
	})
	require.NoError(t, err)

	// GetAllGroups should return distinct Group objects (unique by group ID)
	groups, err := repo.GetAllGroups(ctx, subjectID, subjectType)
	require.NoError(t, err)
	assert.Equal(t, 3, len(groups)) // Should be 3 unique groups, not 4 roles

	// Verify Group structure (not Role)
	for _, group := range groups {
		assert.NotZero(t, group.ID)
		assert.NotZero(t, group.UID)
		assert.NotZero(t, group.Name)
		assert.NotZero(t, group.Description)
		assert.NotZero(t, group.CreatedAt)
		assert.NotZero(t, group.UpdatedAt)
	}

	// Verify order by UID
	for i := 1; i < len(groups); i++ {
		assert.True(t, groups[i].UID >= groups[i-1].UID, "Groups should be sorted by UID")
	}
}

func TestSubjectRepository_GetAllGroups_Empty(t *testing.T) {
	db := setupIntegrationTest(t)
	ctx := context.Background()
	repo := repository.NewSubjectRepository(db)

	subjectID := "user-nonexistent"
	subjectType := "user"

	groups, err := repo.GetAllGroups(ctx, subjectID, subjectType)
	require.NoError(t, err)
	assert.Equal(t, 0, len(groups))
}
