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

func TestAdapter_GroupRepository_Create(t *testing.T) {
	tests := []struct {
		name              string
		data              map[string]any
		mockPgExpectation func(mockPool *pgxmock.ExpectedQuery)
		wantErr           bool
		errMsg            string
	}{
		{
			name: "Happy Path - Create valid group",
			data: map[string]any{
				"uid":         "group-uid-001",
				"name":        "group-name-001",
				"description": "group-description-001",
			},
			mockPgExpectation: func(expectedQuery *pgxmock.ExpectedQuery) {
				rows := pgxmock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow(int64(1), time.Time{}, time.Time{})
				expectedQuery.WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "Error - Unique constraint violation on name",
			data: map[string]any{
				"uid":         "group-uid-002",
				"name":        "existing-group",
				"description": "group-description-002",
			},
			mockPgExpectation: func(expectedQuery *pgxmock.ExpectedQuery) {
				expectedQuery.WillReturnError(&pgconn.PgError{
					ConstraintName: "uq_group_name",
				})
			},
			wantErr: true,
			errMsg:  "already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			repo := NewGroupRepository(mockPool)
			group := &model.Group{
				UID:         tt.data["uid"].(string),
				Name:        tt.data["name"].(string),
				Description: tt.data["description"].(string),
			}

			if tt.mockPgExpectation != nil {
				tt.mockPgExpectation(
					mockPool.ExpectQuery(`
							INSERT INTO "group" \(uid, name, description\)
							VALUES \(\$1, \$2, \$3\)
							RETURNING id, created_at, updated_at
						`).
						WithArgs(group.UID, group.Name, group.Description))
			}
			err = repo.Create(context.Background(), group)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.data["uid"], group.UID)
			assert.Equal(t, tt.data["name"], group.Name)
			assert.Equal(t, tt.data["description"], group.Description)
			assert.Equal(t, int64(1), group.ID)
			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestAdapter_GroupRepository_Create_EmptyUID(t *testing.T) {
	mockDB, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockDB.Close()

	repo := NewGroupRepository(mockDB)

	group := &model.Group{
		Name:        "Test Group",
		Description: "Test Description",
		UID:         "", // Empty UID
	}

	err = repo.Create(context.Background(), group)

	assert.ErrorIs(t, err, errors.ErrInvalidEntity)
}

func TestAdapter_GroupRepository_Update(t *testing.T) {
	tests := []struct {
		name    string
		data    map[string]any
		wantErr bool
		errMsg  string
	}{
		{
			name: "Happy Path - Update existing group",
			data: map[string]any{
				"id":          int64(1),
				"uid":         "group-uid-001",
				"name":        "updated-group",
				"description": "updated-description",
			},
			wantErr: false,
		},
		{
			name: "Error - Unique constraint violation",
			data: map[string]any{
				"id":          int64(1),
				"uid":         "group-uid-002",
				"name":        "existing-group",
				"description": "description-002",
			},
			wantErr: true,
			errMsg:  "already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			repo := NewGroupRepository(mockPool)
			group := &model.Group{
				ID:          tt.data["id"].(int64),
				UID:         tt.data["uid"].(string),
				Name:        tt.data["name"].(string),
				Description: tt.data["description"].(string),
			}

			if tt.wantErr && tt.errMsg == "already exists" {
				mockPool.ExpectExec(`
					UPDATE "group"
					SET name = \$2, description = \$3, updated_at = NOW\(\)
					WHERE id = \$1
				`).
					WithArgs(group.ID, group.Name, group.Description).
					WillReturnError(&pgconn.PgError{
						ConstraintName: "uq_group_name",
					})
			} else {
				mockPool.ExpectExec(`
					UPDATE "group"
					SET name = \$2, description = \$3, updated_at = NOW\(\)
					WHERE id = \$1
				`).
					WithArgs(group.ID, group.Name, group.Description).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			}

			err = repo.Update(context.Background(), group)

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

func TestAdapter_GroupRepository_Delete(t *testing.T) {
	tests := []struct {
		name    string
		data    map[string]any
		wantErr bool
		errMsg  string
	}{
		{
			name: "Happy Path - Delete existing group",
			data: map[string]any{
				"id": int64(1),
			},
			wantErr: false,
		},
		{
			name: "Error - Group not found",
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

			repo := NewGroupRepository(mockPool)
			id := tt.data["id"].(int64)

			if tt.wantErr {
				mockPool.ExpectExec(`DELETE FROM "group" WHERE id = \$1`).
					WithArgs(id).
					WillReturnResult(pgxmock.NewResult("DELETE", 0))
			} else {
				mockPool.ExpectExec(`DELETE FROM "group" WHERE id = \$1`).
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

func TestAdapter_GroupRepository_GetByID(t *testing.T) {
	tests := []struct {
		name    string
		data    map[string]any
		wantErr bool
		errMsg  string
	}{
		{
			name: "Happy Path - Get existing group",
			data: map[string]any{
				"id":   int64(1),
				"uid":  "group-uid-001",
				"name": "group-name-001",
			},
			wantErr: false,
		},
		{
			name: "Error - Group not found",
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

			repo := NewGroupRepository(mockPool)
			id := tt.data["id"].(int64)

			if tt.wantErr {
				mockPool.ExpectQuery(`
					SELECT id, uid, name, description, created_at, updated_at
					FROM "group"
					WHERE id = \$1
				`).
					WithArgs(id).
					WillReturnError(pgx.ErrNoRows)
			} else {
				rows := pgxmock.NewRows([]string{"id", "uid", "name", "description", "created_at", "updated_at"}).
					AddRow(tt.data["id"], tt.data["uid"], tt.data["name"], "description", time.Time{}, time.Time{})
				mockPool.ExpectQuery(`
					SELECT id, uid, name, description, created_at, updated_at
					FROM "group"
					WHERE id = \$1
				`).
					WithArgs(id).
					WillReturnRows(rows)
			}

			group, err := repo.GetByID(context.Background(), id)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, group)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, group)
			assert.Equal(t, tt.data["id"], group.ID)
			assert.Equal(t, tt.data["uid"], group.UID)
			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestAdapter_GroupRepository_List(t *testing.T) {
	tests := []struct {
		name       string
		pagination *param.PaginationParam
		filter     *param.GroupListFilterParam
		wantErr    bool
		wantCount  int
		wantTotal  int64
	}{
		{
			name: "Happy Path - List all groups",
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
			name: "Happy Path - List with Name filter",
			pagination: &param.PaginationParam{
				Page:  func() *int { i := 1; return &i }(),
				Limit: func() *int { i := 10; return &i }(),
			},
			filter: &param.GroupListFilterParam{
				Name: func() *string { s := "group-001"; return &s }(),
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

			repo := NewGroupRepository(mockPool)

			// Calculate expected arguments based on filter
			filterArgCount := 0
			if tt.filter != nil {
				if len(tt.filter.IDs) > 0 {
					filterArgCount += 1
				}
				if len(tt.filter.UIDs) > 0 {
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
			dataRows := pgxmock.NewRows([]string{"id", "uid", "name", "description", "created_at", "updated_at"})
			for i := 0; i < tt.wantCount; i++ {
				dataRows.AddRow(int64(i+1), fmt.Sprintf("uid-%d", i+1), fmt.Sprintf("group-%d", i+1), "description", time.Time{}, time.Time{})
			}
			mockPool.ExpectQuery(`SELECT id, uid, name, description, created_at, updated_at FROM "group" WHERE 1=1`).
				WithArgs(dataAnyArgs...).
				WillReturnRows(dataRows)

			// Set up count expectations
			countRows := pgxmock.NewRows([]string{"count"}).AddRow(tt.wantTotal)
			mockPool.ExpectQuery(`SELECT COUNT\(\*\) FROM "group" WHERE 1=1`).
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

func TestAdapter_GroupRepository_ListPermission(t *testing.T) {
	tests := []struct {
		name       string
		groupID    int64
		pagination *param.PaginationParam
		filter     *param.GroupPermissionListFilterParam
		wantErr    bool
		wantCount  int
		wantTotal  int64
	}{
		{
			name:       "Happy Path - List all permissions",
			groupID:    1,
			pagination: &param.PaginationParam{Page: func() *int { i := 1; return &i }(), Limit: func() *int { i := 10; return &i }()},
			filter:     nil,
			wantErr:    false,
			wantCount:  2,
			wantTotal:  2,
		},
		{
			name:       "Happy Path - Empty list",
			groupID:    1,
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

			repo := NewGroupRepository(mockPool)

			// Calculate expected arguments based on filter
			filterArgCount := 1 // groupID is always first
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

			// Set up data query expectations
			dataRows := pgxmock.NewRows([]string{"id", "uid", "group_id", "group_uid", "permission_id", "permission_uid", "resource", "action", "description", "created_at"})
			for i := 0; i < tt.wantCount; i++ {
				dataRows.AddRow(int64(1), "gp-uid", int64(1), "group-uid", int64(i+1), fmt.Sprintf("perm-uid-%d", i+1), fmt.Sprintf("resource-%d", i+1), fmt.Sprintf("action-%d", i+1), "description", time.Time{})
			}
			mockPool.ExpectQuery(`SELECT gp\.id, gp\.uid, gp\.group_id, g\.uid as group_uid, gp\.permission_id, p\.uid as permission_uid, p\.resource, p\.action, p\.description, gp\.created_at FROM group_permission gp JOIN "group" g ON gp\.group_id = g\.id JOIN permission p ON gp\.permission_id = p\.id WHERE gp\.group_id = \$1`).
				WithArgs(dataAnyArgs...).
				WillReturnRows(dataRows)

			// Set up count expectations
			countRows := pgxmock.NewRows([]string{"count"}).AddRow(tt.wantTotal)
			mockPool.ExpectQuery(`SELECT COUNT\(\*\) FROM group_permission gp JOIN permission p ON gp\.permission_id = p\.id WHERE gp\.group_id = \$1`).
				WithArgs(countAnyArgs...).
				WillReturnRows(countRows)

			result, err := repo.ListPermission(context.Background(), tt.groupID, tt.pagination, tt.filter)

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

func TestAdapter_GroupRepository_AddPermission(t *testing.T) {
	tests := []struct {
		name         string
		groupID      int64
		permissionID int64
		wantErr      bool
		errMsg       string
	}{
		{
			name:         "Happy Path - Add permission to group",
			groupID:      1,
			permissionID: 1,
			wantErr:      false,
		},
		{
			name:         "Error - Group not found (foreign key)",
			groupID:      999,
			permissionID: 1,
			wantErr:      true,
			errMsg:       "not found",
		},
		{
			name:         "Error - Permission not found (foreign key)",
			groupID:      1,
			permissionID: 999,
			wantErr:      true,
			errMsg:       "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			repo := NewGroupRepository(mockPool)

			if tt.wantErr && tt.groupID == 999 {
				mockPool.ExpectExec(`
					INSERT INTO group_permission \(uid, group_id, permission_id\)
					VALUES \(\$1, \$2, \$3\)
					ON CONFLICT \(group_id, permission_id\) DO NOTHING
				`).
					WithArgs(pgxmock.AnyArg(), tt.groupID, tt.permissionID).
					WillReturnError(&pgconn.PgError{
						ConstraintName: "fk_group_permission_group",
					})
			} else if tt.wantErr && tt.permissionID == 999 {
				mockPool.ExpectExec(`
					INSERT INTO group_permission \(uid, group_id, permission_id\)
					VALUES \(\$1, \$2, \$3\)
					ON CONFLICT \(group_id, permission_id\) DO NOTHING
				`).
					WithArgs(pgxmock.AnyArg(), tt.groupID, tt.permissionID).
					WillReturnError(&pgconn.PgError{
						ConstraintName: "fk_group_permission_permission",
					})
			} else {
				mockPool.ExpectExec(`
					INSERT INTO group_permission \(uid, group_id, permission_id\)
					VALUES \(\$1, \$2, \$3\)
					ON CONFLICT \(group_id, permission_id\) DO NOTHING
				`).
					WithArgs(pgxmock.AnyArg(), tt.groupID, tt.permissionID).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			}

			err = repo.AddPermission(context.Background(), tt.groupID, tt.permissionID)

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

func TestAdapter_GroupRepository_RemovePermission(t *testing.T) {
	tests := []struct {
		name         string
		groupID      int64
		permissionID int64
		wantErr      bool
		errMsg       string
	}{
		{
			name:         "Happy Path - Remove permission from group",
			groupID:      1,
			permissionID: 1,
			wantErr:      false,
		},
		{
			name:         "Error - Permission not found in group",
			groupID:      1,
			permissionID: 999,
			wantErr:      true,
			errMsg:       "not found in group",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			repo := NewGroupRepository(mockPool)

			if tt.wantErr {
				mockPool.ExpectExec(`DELETE FROM group_permission WHERE group_id = \$1 AND permission_id = \$2`).
					WithArgs(tt.groupID, tt.permissionID).
					WillReturnResult(pgxmock.NewResult("DELETE", 0))
			} else {
				mockPool.ExpectExec(`DELETE FROM group_permission WHERE group_id = \$1 AND permission_id = \$2`).
					WithArgs(tt.groupID, tt.permissionID).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
			}

			err = repo.RemovePermission(context.Background(), tt.groupID, tt.permissionID)

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

func TestAdapter_GroupRepository_ReplacePermission(t *testing.T) {
	tests := []struct {
		name          string
		groupID       int64
		permissionIDs []int64
		wantErr       bool
		errMsg        string
	}{
		{
			name:          "Happy Path - Replace permissions",
			groupID:       1,
			permissionIDs: []int64{1, 2, 3},
			wantErr:       false,
		},
		{
			name:          "Happy Path - Replace with empty permissions",
			groupID:       1,
			permissionIDs: []int64{},
			wantErr:       false,
		},
		{
			name:          "Happy Path - Replace with single permission",
			groupID:       1,
			permissionIDs: []int64{1},
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For ReplacePermission, we need to mock a transaction (pgxmock.NewConn)
			mockDB, err := pgxmock.NewConn(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherRegexp))
			require.NoError(t, err)
			defer mockDB.Close(context.Background())

			repo := NewGroupRepository(mockDB)

			// Mock delete
			mockDB.ExpectExec(`DELETE FROM group_permission WHERE group_id = \$1`).
				WithArgs(tt.groupID).
				WillReturnResult(pgxmock.NewResult("DELETE", 0))

			// Mock inserts for each permission
			for _, permID := range tt.permissionIDs {
				mockDB.ExpectExec(`INSERT INTO group_permission \(uid, group_id, permission_id\) VALUES \(\$1, \$2, \$3\)`).
					WithArgs(pgxmock.AnyArg(), tt.groupID, permID).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			}

			err = repo.ReplacePermission(context.Background(), tt.groupID, tt.permissionIDs)

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
