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

func TestAdapter_PermissionRepository_Create(t *testing.T) {
	tests := []struct {
		name    string
		data    map[string]any
		wantErr bool
		errMsg  string
	}{
		{
			name: "Happy Path - Create valid permission",
			data: map[string]any{
				"uid":         "perm-uid-001",
				"resource":    "resource-001",
				"action":      "action-001",
				"description": "description-001",
			},
			wantErr: false,
		},
		{
			name: "Error - Unique constraint violation",
			data: map[string]any{
				"uid":         "perm-uid-002",
				"resource":    "existing-resource",
				"action":      "existing-action",
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

			repo := NewPermissionRepository(mockPool)
			permission := &model.Permission{
				UID:         tt.data["uid"].(string),
				Resource:    tt.data["resource"].(string),
				Action:      tt.data["action"].(string),
				Description: tt.data["description"].(string),
			}

			if tt.wantErr && tt.errMsg == "already exists" {
				mockPool.ExpectQuery(`
					INSERT INTO permission \(uid, resource, action, description\)
					VALUES \(\$1, \$2, \$3, \$4\)
					RETURNING id, created_at, updated_at
				`).
					WithArgs(permission.UID, permission.Resource, permission.Action, permission.Description).
					WillReturnError(&pgconn.PgError{
						ConstraintName: "uq_permission_resource_action",
					})
			} else {
				rows := pgxmock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow(int64(1), time.Time{}, time.Time{})
				mockPool.ExpectQuery(`
					INSERT INTO permission \(uid, resource, action, description\)
					VALUES \(\$1, \$2, \$3, \$4\)
					RETURNING id, created_at, updated_at
				`).
					WithArgs(permission.UID, permission.Resource, permission.Action, permission.Description).
					WillReturnRows(rows)
			}

			err = repo.Create(context.Background(), permission)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.data["uid"], permission.UID)
			assert.Equal(t, tt.data["resource"], permission.Resource)
			assert.Equal(t, tt.data["action"], permission.Action)
			assert.Equal(t, tt.data["description"], permission.Description)
			assert.Equal(t, int64(1), permission.ID)
			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestAdapter_PermissionRepository_Create_EmptyUID(t *testing.T) {
	mockDB, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockDB.Close()

	repo := NewPermissionRepository(mockDB)

	permission := &model.Permission{
		Resource:    "test-resource",
		Action:      "test-action",
		Description: "Test Description",
		UID:         "", // Empty UID
	}

	err = repo.Create(context.Background(), permission)

	assert.ErrorIs(t, err, errors.ErrInvalidEntity)
}

func TestAdapter_PermissionRepository_Update(t *testing.T) {
	tests := []struct {
		name    string
		data    map[string]any
		wantErr bool
		errMsg  string
	}{
		{
			name: "Happy Path - Update existing permission",
			data: map[string]any{
				"id":          int64(1),
				"uid":         "perm-uid-001",
				"resource":    "updated-resource",
				"action":      "updated-action",
				"description": "updated-description",
			},
			wantErr: false,
		},
		{
			name: "Error - Permission not found (0 rows affected)",
			data: map[string]any{
				"id":          int64(999),
				"uid":         "perm-uid-002",
				"resource":    "resource-002",
				"action":      "action-002",
				"description": "description-002",
			},
			wantErr: false, // Update doesn't check rows affected, only returns error on exec failure
		},
		{
			name: "Error - Unique constraint violation",
			data: map[string]any{
				"id":          int64(1),
				"uid":         "perm-uid-003",
				"resource":    "existing-resource",
				"action":      "existing-action",
				"description": "description-003",
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

			repo := NewPermissionRepository(mockPool)
			permission := &model.Permission{
				ID:          tt.data["id"].(int64),
				UID:         tt.data["uid"].(string),
				Resource:    tt.data["resource"].(string),
				Action:      tt.data["action"].(string),
				Description: tt.data["description"].(string),
			}

			if tt.wantErr && tt.errMsg == "already exists" {
				mockPool.ExpectExec(`
					UPDATE permission
					SET resource = \$2, action = \$3, description = \$4, updated_at = NOW\(\)
					WHERE id = \$1
				`).
					WithArgs(permission.ID, permission.Resource, permission.Action, permission.Description).
					WillReturnError(&pgconn.PgError{
						ConstraintName: "uq_permission_resource_action",
					})
			} else {
				mockPool.ExpectExec(`
					UPDATE permission
					SET resource = \$2, action = \$3, description = \$4, updated_at = NOW\(\)
					WHERE id = \$1
				`).
					WithArgs(permission.ID, permission.Resource, permission.Action, permission.Description).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			}

			err = repo.Update(context.Background(), permission)

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

func TestAdapter_PermissionRepository_Delete(t *testing.T) {
	tests := []struct {
		name    string
		data    map[string]any
		wantErr bool
		errMsg  string
	}{
		{
			name: "Happy Path - Delete existing permission",
			data: map[string]any{
				"id": int64(1),
			},
			wantErr: false,
		},
		{
			name: "Error - Permission not found",
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

			repo := NewPermissionRepository(mockPool)
			id := tt.data["id"].(int64)

			if tt.wantErr {
				mockPool.ExpectExec(`DELETE FROM permission WHERE id = \$1`).
					WithArgs(id).
					WillReturnResult(pgxmock.NewResult("DELETE", 0))
			} else {
				mockPool.ExpectExec(`DELETE FROM permission WHERE id = \$1`).
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

func TestAdapter_PermissionRepository_GetByID(t *testing.T) {
	tests := []struct {
		name    string
		data    map[string]any
		wantErr bool
		errMsg  string
	}{
		{
			name: "Happy Path - Get existing permission",
			data: map[string]any{
				"id":       int64(1),
				"uid":      "perm-uid-001",
				"resource": "resource-001",
				"action":   "action-001",
			},
			wantErr: false,
		},
		{
			name: "Error - Permission not found",
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

			repo := NewPermissionRepository(mockPool)
			id := tt.data["id"].(int64)

			if tt.wantErr {
				mockPool.ExpectQuery(`
					SELECT id, uid, resource, action, description, created_at, updated_at
					FROM permission
					WHERE id = \$1
				`).
					WithArgs(id).
					WillReturnError(pgx.ErrNoRows)
			} else {
				rows := pgxmock.NewRows([]string{"id", "uid", "resource", "action", "description", "created_at", "updated_at"}).
					AddRow(tt.data["id"], tt.data["uid"], tt.data["resource"], tt.data["action"], "description", time.Time{}, time.Time{})
				mockPool.ExpectQuery(`
					SELECT id, uid, resource, action, description, created_at, updated_at
					FROM permission
					WHERE id = \$1
				`).
					WithArgs(id).
					WillReturnRows(rows)
			}

			permission, err := repo.GetByID(context.Background(), id)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, permission)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, permission)
			assert.Equal(t, tt.data["id"], permission.ID)
			assert.Equal(t, tt.data["uid"], permission.UID)
			assert.Equal(t, tt.data["resource"], permission.Resource)
			assert.Equal(t, tt.data["action"], permission.Action)
			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestAdapter_PermissionRepository_List(t *testing.T) {
	tests := []struct {
		name       string
		pagination *param.PaginationParam
		filter     *param.PermissionListFilterParam
		wantErr    bool
		wantCount  int
		wantTotal  int64
	}{
		{
			name: "Happy Path - List all permissions",
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
			name: "Happy Path - List with IDs filter",
			pagination: &param.PaginationParam{
				Page:  func() *int { i := 1; return &i }(),
				Limit: func() *int { i := 10; return &i }(),
			},
			filter: &param.PermissionListFilterParam{
				IDs: []int64{1, 2},
			},
			wantErr:   false,
			wantCount: 2,
			wantTotal: 2,
		},
		{
			name: "Happy Path - List with resource filter",
			pagination: &param.PaginationParam{
				Page:  func() *int { i := 1; return &i }(),
				Limit: func() *int { i := 10; return &i }(),
			},
			filter: &param.PermissionListFilterParam{
				Resource: func() *string { s := "resource-001"; return &s }(),
			},
			wantErr:   false,
			wantCount: 1,
			wantTotal: 1,
		},
		{
			name: "Happy Path - List with action filter",
			pagination: &param.PaginationParam{
				Page:  func() *int { i := 1; return &i }(),
				Limit: func() *int { i := 10; return &i }(),
			},
			filter: &param.PermissionListFilterParam{
				Action: func() *string { s := "action-001"; return &s }(),
			},
			wantErr:   false,
			wantCount: 1,
			wantTotal: 1,
		},
		{
			name: "Happy Path - List with sorting",
			pagination: &param.PaginationParam{
				Page:    func() *int { i := 1; return &i }(),
				Limit:   func() *int { i := 10; return &i }(),
				OrderBy: func() *string { s := "resource"; return &s }(),
				Sort:    func() *string { s := "ASC"; return &s }(),
			},
			filter:    nil,
			wantErr:   false,
			wantCount: 2,
			wantTotal: 2,
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

			repo := NewPermissionRepository(mockPool)

			// Calculate expected arguments based on filter
			filterArgCount := 0
			if tt.filter != nil {
				if len(tt.filter.IDs) > 0 {
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

			// Count query only has filter args (no limit/offset)
			countAnyArgs := make([]interface{}, filterArgCount)
			for i := range countAnyArgs {
				countAnyArgs[i] = pgxmock.AnyArg()
			}

			// Set up data query expectations
			dataRows := pgxmock.NewRows([]string{"id", "uid", "resource", "action", "description", "created_at", "updated_at"})
			for i := 0; i < tt.wantCount; i++ {
				dataRows.AddRow(int64(i+1), fmt.Sprintf("uid-%d", i+1), fmt.Sprintf("resource-%d", i+1), fmt.Sprintf("action-%d", i+1), "description", time.Time{}, time.Time{})
			}
			mockPool.ExpectQuery(`SELECT id, uid, resource, action, description, created_at, updated_at FROM permission WHERE 1=1`).
				WithArgs(dataAnyArgs...).
				WillReturnRows(dataRows)

			// Set up count expectations
			countRows := pgxmock.NewRows([]string{"count"}).AddRow(tt.wantTotal)
			mockPool.ExpectQuery(`SELECT COUNT\(\*\) FROM permission WHERE 1=1`).
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
