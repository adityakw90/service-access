package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/adityakw90/service-access/internal/core/domain/errors"
	"github.com/adityakw90/service-access/internal/core/domain/model"
	"github.com/adityakw90/service-access/internal/core/domain/param"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdapter_RoleRepository_Create(t *testing.T) {
	tests := []struct {
		name    string
		data    map[string]any
		wantErr bool
		errMsg  string
	}{
		{
			name: "Happy Path - Create valid role",
			data: map[string]any{
				"uid":         "role-uid-001",
				"group_id":    int64(1),
				"name":        "role-name-001",
				"description": "description-001",
			},
			wantErr: false,
		},
		{
			name: "Error - Unique constraint violation on (group_id, name)",
			data: map[string]any{
				"uid":         "role-uid-002",
				"group_id":    int64(1),
				"name":        "existing-role",
				"description": "description-002",
			},
			wantErr: true,
			errMsg:  "already exists in group",
		},
		{
			name: "Error - Foreign key violation on group_id",
			data: map[string]any{
				"uid":         "role-uid-003",
				"group_id":    int64(999),
				"name":        "role-name-003",
				"description": "description-003",
			},
			wantErr: true,
			errMsg:  "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			repo := NewRoleRepository(mockPool)
			role := &model.Role{
				UID:         tt.data["uid"].(string),
				GroupID:     tt.data["group_id"].(int64),
				Name:        tt.data["name"].(string),
				Description: tt.data["description"].(string),
			}

			if tt.wantErr && tt.errMsg == "already exists in group" {
				mockPool.ExpectQuery(`
					INSERT INTO role \(uid, group_id, name, description\)
					VALUES \(\$1, \$2, \$3, \$4\)
					RETURNING id, created_at, updated_at
				`).
					WithArgs(role.UID, role.GroupID, role.Name, role.Description).
					WillReturnError(&pgconn.PgError{
						ConstraintName: "uq_role_group_name",
					})
			} else if tt.wantErr && tt.errMsg == "not found" {
				mockPool.ExpectQuery(`
					INSERT INTO role \(uid, group_id, name, description\)
					VALUES \(\$1, \$2, \$3, \$4\)
					RETURNING id, created_at, updated_at
				`).
					WithArgs(role.UID, role.GroupID, role.Name, role.Description).
					WillReturnError(&pgconn.PgError{
						ConstraintName: "fk_role_group",
					})
			} else {
				rows := pgxmock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow(int64(1), time.Time{}, time.Time{})
				mockPool.ExpectQuery(`
					INSERT INTO role \(uid, group_id, name, description\)
					VALUES \(\$1, \$2, \$3, \$4\)
					RETURNING id, created_at, updated_at
				`).
					WithArgs(role.UID, role.GroupID, role.Name, role.Description).
					WillReturnRows(rows)
			}

			err = repo.Create(context.Background(), role)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.data["uid"], role.UID)
			assert.Equal(t, tt.data["group_id"], role.GroupID)
			assert.Equal(t, tt.data["name"], role.Name)
			assert.Equal(t, int64(1), role.ID)
			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestAdapter_RoleRepository_Create_EmptyUID(t *testing.T) {
	mockDB, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockDB.Close()

	repo := NewRoleRepository(mockDB)

	role := &model.Role{
		GroupID:     1,
		Name:        "Test Role",
		Description: "Test Description",
		UID:         "", // Empty UID
	}

	err = repo.Create(context.Background(), role)

	assert.ErrorIs(t, err, errors.ErrInvalidEntity)
}

func TestAdapter_RoleRepository_Update(t *testing.T) {
	tests := []struct {
		name    string
		data    map[string]any
		wantErr bool
		errMsg  string
	}{
		{
			name: "Happy Path - Update existing role",
			data: map[string]any{
				"id":          int64(1),
				"uid":         "role-uid-001",
				"group_id":    int64(1),
				"name":        "updated-role",
				"description": "updated-description",
			},
			wantErr: false,
		},
		{
			name: "Error - Unique constraint violation",
			data: map[string]any{
				"id":          int64(1),
				"uid":         "role-uid-002",
				"group_id":    int64(1),
				"name":        "existing-role",
				"description": "description-002",
			},
			wantErr: true,
			errMsg:  "already exists in group",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			repo := NewRoleRepository(mockPool)
			role := &model.Role{
				ID:          tt.data["id"].(int64),
				UID:         tt.data["uid"].(string),
				GroupID:     tt.data["group_id"].(int64),
				Name:        tt.data["name"].(string),
				Description: tt.data["description"].(string),
			}

			if tt.wantErr && tt.errMsg == "already exists in group" {
				mockPool.ExpectExec(`
					UPDATE role
					SET name = \$2, description = \$3, updated_at = NOW\(\)
					WHERE id = \$1
				`).
					WithArgs(role.ID, role.Name, role.Description).
					WillReturnError(&pgconn.PgError{
						ConstraintName: "uq_role_group_name",
					})
			} else {
				mockPool.ExpectExec(`
					UPDATE role
					SET name = \$2, description = \$3, updated_at = NOW\(\)
					WHERE id = \$1
				`).
					WithArgs(role.ID, role.Name, role.Description).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			}

			err = repo.Update(context.Background(), role)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			assert.NoError(t, err)
			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestAdapter_RoleRepository_Delete(t *testing.T) {
	tests := []struct {
		name    string
		data    map[string]any
		wantErr bool
		errMsg  string
	}{
		{
			name: "Happy Path - Delete existing role",
			data: map[string]any{
				"id": int64(1),
			},
			wantErr: false,
		},
		{
			name: "Error - Role not found",
			data: map[string]any{
				"id": int64(999),
			},
			wantErr: true,
			errMsg:  "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			repo := NewRoleRepository(mockPool)
			id := tt.data["id"].(int64)

			if tt.wantErr {
				mockPool.ExpectExec(`DELETE FROM role WHERE id = \$1`).
					WithArgs(id).
					WillReturnResult(pgxmock.NewResult("DELETE", 0))
			} else {
				mockPool.ExpectExec(`DELETE FROM role WHERE id = \$1`).
					WithArgs(id).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
			}

			err = repo.Delete(context.Background(), id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			assert.NoError(t, err)
			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestAdapter_RoleRepository_GetByID(t *testing.T) {
	tests := []struct {
		name    string
		data    map[string]any
		wantErr bool
		errMsg  string
	}{
		{
			name: "Happy Path - Get existing role",
			data: map[string]any{
				"id":        int64(1),
				"uid":       "role-uid-001",
				"group_id":  int64(1),
				"group_uid": "group-uid-001",
				"name":      "role-name-001",
			},
			wantErr: false,
		},
		{
			name: "Error - Role not found",
			data: map[string]any{
				"id": int64(999),
			},
			wantErr: true,
			errMsg:  "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			repo := NewRoleRepository(mockPool)
			id := tt.data["id"].(int64)

			if tt.wantErr {
				mockPool.ExpectQuery(`
					SELECT r\.id, r\.uid, r\.group_id, g\.uid as group_uid, r\.name, r\.description, r\.created_at, r\.updated_at
					FROM role r
					JOIN "group" g ON r\.group_id = g\.id
					WHERE r\.id = \$1
				`).
					WithArgs(id).
					WillReturnError(pgx.ErrNoRows)
			} else {
				rows := pgxmock.NewRows([]string{"id", "uid", "group_id", "group_uid", "name", "description", "created_at", "updated_at"}).
					AddRow(tt.data["id"], tt.data["uid"], tt.data["group_id"], tt.data["group_uid"], tt.data["name"], "description", time.Time{}, time.Time{})
				mockPool.ExpectQuery(`
					SELECT r\.id, r\.uid, r\.group_id, g\.uid as group_uid, r\.name, r\.description, r\.created_at, r\.updated_at
					FROM role r
					JOIN "group" g ON r\.group_id = g\.id
					WHERE r\.id = \$1
				`).
					WithArgs(id).
					WillReturnRows(rows)
			}

			role, err := repo.GetByID(context.Background(), id)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, role)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, role)
			assert.Equal(t, tt.data["id"], role.ID)
			assert.Equal(t, tt.data["uid"], role.UID)
			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestAdapter_RoleRepository_List(t *testing.T) {
	tests := []struct {
		name       string
		pagination *param.PaginationParam
		filter     *param.RoleListFilterParam
		wantErr    bool
		wantCount  int
		wantTotal  int64
	}{
		{
			name: "Happy Path - List all roles",
			pagination: &param.PaginationParam{
				Page:  func() *int { i := 1; return &i }(),
				Limit: func() *int { i := 10; return &i }(),
			},
			filter:    nil,
			wantErr:   false,
			wantCount: 2,
			wantTotal: 2,
		},
		{
			name: "Happy Path - List with GroupID filter",
			pagination: &param.PaginationParam{
				Page:  func() *int { i := 1; return &i }(),
				Limit: func() *int { i := 10; return &i }(),
			},
			filter: &param.RoleListFilterParam{
				GroupID: func() *int64 { i := int64(1); return &i }(),
			},
			wantErr:   false,
			wantCount: 1,
			wantTotal: 1,
		},
		{
			name: "Happy Path - Empty list",
			pagination: &param.PaginationParam{
				Page:  func() *int { i := 1; return &i }(),
				Limit: func() *int { i := 10; return &i }(),
			},
			filter:    nil,
			wantErr:   false,
			wantCount: 0,
			wantTotal: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherRegexp))
			require.NoError(t, err)
			defer mockPool.Close()

			repo := NewRoleRepository(mockPool)

			// Calculate expected arguments based on filter
			filterArgCount := 0
			if tt.filter != nil {
				if len(tt.filter.IDs) > 0 {
					filterArgCount += 1
				}
				if len(tt.filter.UIDs) > 0 {
					filterArgCount += 1
				}
				if tt.filter.GroupID != nil {
					filterArgCount += 1
				}
				if tt.filter.GroupUID != nil {
					filterArgCount += 1
				}
				if tt.filter.Name != nil {
					filterArgCount += 1
				}
				if tt.filter.Query != nil {
					filterArgCount += 2
				}
			}

			// Data query has filter args + limit + offset
			dataArgCount := filterArgCount + 2
			dataAnyArgs := make([]interface{}, dataArgCount)
			for i := range dataAnyArgs {
				dataAnyArgs[i] = pgxmock.AnyArg()
			}

			// Count query only has filter args
			countAnyArgs := make([]interface{}, filterArgCount)
			for i := range countAnyArgs {
				countAnyArgs[i] = pgxmock.AnyArg()
			}

			// Set up data query expectations
			dataRows := pgxmock.NewRows([]string{"id", "uid", "group_id", "group_uid", "name", "description", "created_at", "updated_at"})
			for i := 0; i < tt.wantCount; i++ {
				dataRows.AddRow(int64(i+1), fmt.Sprintf("uid-%d", i+1), int64(1), "group-uid", fmt.Sprintf("role-%d", i+1), "description", time.Time{}, time.Time{})
			}
			mockPool.ExpectQuery(`SELECT r\.id, r\.uid, r\.group_id, g\.uid as group_uid, r\.name, r\.description, r\.created_at, r\.updated_at FROM role r JOIN "group" g ON r\.group_id = g\.id WHERE 1=1`).
				WithArgs(dataAnyArgs...).
				WillReturnRows(dataRows)

			// Set up count expectations
			countRows := pgxmock.NewRows([]string{"count"}).AddRow(tt.wantTotal)
			mockPool.ExpectQuery(`SELECT COUNT\(\*\) FROM role r JOIN "group" g ON r\.group_id = g\.id WHERE 1=1`).
				WithArgs(countAnyArgs...).
				WillReturnRows(countRows)

			result, err := repo.List(context.Background(), tt.pagination, tt.filter)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Len(t, result.Items, tt.wantCount)
			assert.Equal(t, tt.wantTotal, result.Meta.Total)
			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestAdapter_RoleRepository_ListPermission(t *testing.T) {
	tests := []struct {
		name       string
		roleID     int64
		pagination *param.PaginationParam
		filter     *param.RolePermissionListFilterParam
		wantErr    bool
		wantCount  int
		wantTotal  int64
	}{
		{
			name:       "Happy Path - List all permissions",
			roleID:     1,
			pagination: &param.PaginationParam{Page: func() *int { i := 1; return &i }(), Limit: func() *int { i := 10; return &i }()},
			filter:     nil,
			wantErr:    false,
			wantCount:  2,
			wantTotal:  2,
		},
		{
			name:       "Happy Path - Empty list",
			roleID:     1,
			pagination: &param.PaginationParam{Page: func() *int { i := 1; return &i }(), Limit: func() *int { i := 10; return &i }()},
			filter:     nil,
			wantErr:    false,
			wantCount:  0,
			wantTotal:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherRegexp))
			require.NoError(t, err)
			defer mockPool.Close()

			repo := NewRoleRepository(mockPool)

			// Calculate expected arguments based on filter
			filterArgCount := 1 // roleID is always first
			if tt.filter != nil {
				if len(tt.filter.IDs) > 0 {
					filterArgCount += 1
				}
				if len(tt.filter.UIDs) > 0 {
					filterArgCount += 1
				}
				if len(tt.filter.PermissionIDs) > 0 {
					filterArgCount += 1
				}
				if len(tt.filter.PermissionUIDs) > 0 {
					filterArgCount += 1
				}
				if tt.filter.Resource != nil {
					filterArgCount += 1
				}
				if tt.filter.Action != nil {
					filterArgCount += 1
				}
				if tt.filter.Query != nil {
					filterArgCount += 3
				}
			}

			// Data query has filter args + limit + offset
			dataArgCount := filterArgCount + 2
			dataAnyArgs := make([]interface{}, dataArgCount)
			for i := range dataAnyArgs {
				dataAnyArgs[i] = pgxmock.AnyArg()
			}

			// Count query only has filter args
			countAnyArgs := make([]interface{}, filterArgCount)
			for i := range countAnyArgs {
				countAnyArgs[i] = pgxmock.AnyArg()
			}

			// Set up data query expectations - updated to match new SQL with group_permission join
			dataRows := pgxmock.NewRows([]string{"role_id", "role_uid", "group_permission_id", "group_permission_uid", "permission_uid", "resource", "action", "description", "created_at"})
			for i := 0; i < tt.wantCount; i++ {
				dataRows.AddRow(int64(1), "role-uid", int64(i+1), fmt.Sprintf("gp-uid-%d", i+1), fmt.Sprintf("perm-uid-%d", i+1), fmt.Sprintf("resource-%d", i+1), fmt.Sprintf("action-%d", i+1), "description", time.Time{})
			}
			mockPool.ExpectQuery(`SELECT rp\.role_id, r\.uid as role_uid, rp\.group_permission_id, gp\.uid as group_permission_uid, p\.uid as permission_uid, p\.resource, p\.action, p\.description, rp\.created_at FROM role_permission rp JOIN role r ON rp\.role_id = r\.id JOIN group_permission gp ON rp\.group_permission_id = gp\.id JOIN permission p ON gp\.permission_id = p\.id WHERE rp\.role_id = \$1`).
				WithArgs(dataAnyArgs...).
				WillReturnRows(dataRows)

			// Set up count expectations - updated to match new SQL with group_permission join
			countRows := pgxmock.NewRows([]string{"count"}).AddRow(tt.wantTotal)
			mockPool.ExpectQuery(`SELECT COUNT\(\*\) FROM role_permission rp JOIN role r ON rp\.role_id = r\.id JOIN group_permission gp ON rp\.group_permission_id = gp\.id JOIN permission p ON gp\.permission_id = p\.id WHERE rp\.role_id = \$1`).
				WithArgs(countAnyArgs...).
				WillReturnRows(countRows)

			result, err := repo.ListPermission(context.Background(), tt.roleID, tt.pagination, tt.filter)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Len(t, result.Items, tt.wantCount)
			assert.Equal(t, tt.wantTotal, result.Meta.Total)
			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestAdapter_RoleRepository_AddPermission(t *testing.T) {
	tests := []struct {
		name              string
		roleID            int64
		groupPermissionID int64
		wantErr           bool
		errMsg            string
	}{
		{
			name:              "Happy Path - Add permission to role",
			roleID:            1,
			groupPermissionID: 1,
			wantErr:           false,
		},
		{
			name:              "Error - Role not found (foreign key)",
			roleID:            999,
			groupPermissionID: 1,
			wantErr:           true,
			errMsg:            "not found",
		},
		{
			name:              "Error - Group permission not found (foreign key)",
			roleID:            1,
			groupPermissionID: 999,
			wantErr:           true,
			errMsg:            "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			repo := NewRoleRepository(mockPool)

			if tt.wantErr && tt.roleID == 999 {
				mockPool.ExpectExec(`
					INSERT INTO role_permission \(role_id, group_permission_id\)
					VALUES \(\$1, \$2\)
					ON CONFLICT \(role_id, group_permission_id\) DO NOTHING
				`).
					WithArgs(tt.roleID, tt.groupPermissionID).
					WillReturnError(&pgconn.PgError{
						ConstraintName: "fk_role_permission_role",
					})
			} else if tt.wantErr && tt.groupPermissionID == 999 {
				mockPool.ExpectExec(`
					INSERT INTO role_permission \(role_id, group_permission_id\)
					VALUES \(\$1, \$2\)
					ON CONFLICT \(role_id, group_permission_id\) DO NOTHING
				`).
					WithArgs(tt.roleID, tt.groupPermissionID).
					WillReturnError(&pgconn.PgError{
						ConstraintName: "fk_role_permission_group_permission",
					})
			} else {
				mockPool.ExpectExec(`
					INSERT INTO role_permission \(role_id, group_permission_id\)
					VALUES \(\$1, \$2\)
					ON CONFLICT \(role_id, group_permission_id\) DO NOTHING
				`).
					WithArgs(tt.roleID, tt.groupPermissionID).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			}

			err = repo.AddPermission(context.Background(), tt.roleID, tt.groupPermissionID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			assert.NoError(t, err)
			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestAdapter_RoleRepository_RemovePermission(t *testing.T) {
	tests := []struct {
		name              string
		roleID            int64
		groupPermissionID int64
		wantErr           bool
		errMsg            string
	}{
		{
			name:              "Happy Path - Remove permission from role",
			roleID:            1,
			groupPermissionID: 1,
			wantErr:           false,
		},
		{
			name:              "Error - Permission not found in role",
			roleID:            1,
			groupPermissionID: 999,
			wantErr:           true,
			errMsg:            "not found in role",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			repo := NewRoleRepository(mockPool)

			if tt.wantErr {
				mockPool.ExpectExec(`DELETE FROM role_permission WHERE role_id = \$1 AND group_permission_id = \$2`).
					WithArgs(tt.roleID, tt.groupPermissionID).
					WillReturnResult(pgxmock.NewResult("DELETE", 0))
			} else {
				mockPool.ExpectExec(`DELETE FROM role_permission WHERE role_id = \$1 AND group_permission_id = \$2`).
					WithArgs(tt.roleID, tt.groupPermissionID).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
			}

			err = repo.RemovePermission(context.Background(), tt.roleID, tt.groupPermissionID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			assert.NoError(t, err)
			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestAdapter_RoleRepository_ReplacePermission(t *testing.T) {
	tests := []struct {
		name               string
		roleID             int64
		groupPermissionIDs []int64
		wantErr            bool
		errMsg             string
	}{
		{
			name:               "Happy Path - Replace permissions",
			roleID:             1,
			groupPermissionIDs: []int64{1, 2, 3},
			wantErr:            false,
		},
		{
			name:               "Happy Path - Replace with empty permissions",
			roleID:             1,
			groupPermissionIDs: []int64{},
			wantErr:            false,
		},
		{
			name:               "Happy Path - Replace with single permission",
			roleID:             1,
			groupPermissionIDs: []int64{1},
			wantErr:            false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For ReplacePermission, we need to mock a transaction (pgxmock.NewConn)
			mockDB, err := pgxmock.NewConn(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherRegexp))
			require.NoError(t, err)
			defer mockDB.Close(context.Background())

			repo := NewRoleRepository(mockDB)

			// Mock delete
			mockDB.ExpectExec(`DELETE FROM role_permission WHERE role_id = \$1`).
				WithArgs(tt.roleID).
				WillReturnResult(pgxmock.NewResult("DELETE", 0))

			// Mock inserts for each permission
			for _, permID := range tt.groupPermissionIDs {
				mockDB.ExpectExec(`INSERT INTO role_permission \(role_id, group_permission_id\) VALUES \(\$1, \$2\)`).
					WithArgs(tt.roleID, permID).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			}

			err = repo.ReplacePermission(context.Background(), tt.roleID, tt.groupPermissionIDs)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			assert.NoError(t, err)
			assert.NoError(t, mockDB.ExpectationsWereMet())
		})
	}
}
