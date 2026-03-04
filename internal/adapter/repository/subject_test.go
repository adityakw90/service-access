package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/adityakw90/service-access/internal/core/domain/model"
	"github.com/adityakw90/service-access/internal/core/domain/param"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdapter_SubjectRepository_Create(t *testing.T) {
	tests := []struct {
		name    string
		data    map[string]any
		wantErr bool
		errMsg  string
	}{
		{
			name: "Happy Path - Create valid subject role",
			data: map[string]any{
				"subject_id":   "subject-001",
				"subject_type": "user",
				"role_id":      int64(1),
			},
			wantErr: false,
		},
		{
			name: "Error - Unique constraint violation",
			data: map[string]any{
				"subject_id":   "existing-subject",
				"subject_type": "user",
				"role_id":      int64(1),
			},
			wantErr: true,
			errMsg:  "already exists",
		},
		{
			name: "Error - Foreign key violation on role_id",
			data: map[string]any{
				"subject_id":   "subject-002",
				"subject_type": "user",
				"role_id":      int64(999),
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

			repo := NewSubjectRepository(mockPool)
			subject := &model.SubjectRole{
				SubjectID:   tt.data["subject_id"].(string),
				SubjectType: tt.data["subject_type"].(string),
				RoleID:      tt.data["role_id"].(int64),
			}

			if tt.wantErr && tt.errMsg == "already exists" {
				mockPool.ExpectQuery(`
					INSERT INTO subject_role \(subject_id, subject_type, role_id\)
					VALUES \(\$1, \$2, \$3\)
					RETURNING assigned_at
				`).
					WithArgs(subject.SubjectID, subject.SubjectType, subject.RoleID).
					WillReturnError(&pgconn.PgError{
						ConstraintName: "uq_subject_role",
					})
			} else if tt.wantErr && tt.errMsg == "not found" {
				mockPool.ExpectQuery(`
					INSERT INTO subject_role \(subject_id, subject_type, role_id\)
					VALUES \(\$1, \$2, \$3\)
					RETURNING assigned_at
				`).
					WithArgs(subject.SubjectID, subject.SubjectType, subject.RoleID).
					WillReturnError(&pgconn.PgError{
						ConstraintName: "fk_subject_role_role",
					})
			} else {
				rows := pgxmock.NewRows([]string{"assigned_at"}).AddRow(time.Time{})
				mockPool.ExpectQuery(`
					INSERT INTO subject_role \(subject_id, subject_type, role_id\)
					VALUES \(\$1, \$2, \$3\)
					RETURNING assigned_at
				`).
					WithArgs(subject.SubjectID, subject.SubjectType, subject.RoleID).
					WillReturnRows(rows)
			}

			err = repo.Create(context.Background(), subject)

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

func TestAdapter_SubjectRepository_Update(t *testing.T) {
	tests := []struct {
		name    string
		data    map[string]any
		wantErr bool
		errMsg  string
	}{
		{
			name: "Happy Path - Update existing subject role",
			data: map[string]any{
				"subject_id":   "subject-001",
				"subject_type": "user",
				"role_id":      int64(2),
			},
			wantErr: false,
		},
		{
			name: "Error - Foreign key violation on role_id",
			data: map[string]any{
				"subject_id":   "subject-002",
				"subject_type": "user",
				"role_id":      int64(999),
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

			repo := NewSubjectRepository(mockPool)
			subject := &model.SubjectRole{
				SubjectID:   tt.data["subject_id"].(string),
				SubjectType: tt.data["subject_type"].(string),
				RoleID:      tt.data["role_id"].(int64),
			}

			if tt.wantErr && tt.errMsg == "not found" {
				mockPool.ExpectExec(`
					UPDATE subject_role
					SET role_id = \$3
					WHERE subject_id = \$1 AND subject_type = \$2
				`).
					WithArgs(subject.SubjectID, subject.SubjectType, subject.RoleID).
					WillReturnError(&pgconn.PgError{
						ConstraintName: "fk_subject_role_role",
					})
			} else {
				mockPool.ExpectExec(`
					UPDATE subject_role
					SET role_id = \$3
					WHERE subject_id = \$1 AND subject_type = \$2
				`).
					WithArgs(subject.SubjectID, subject.SubjectType, subject.RoleID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			}

			err = repo.Update(context.Background(), subject)

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

func TestAdapter_SubjectRepository_Delete(t *testing.T) {
	tests := []struct {
		name       string
		setupMock  func(pgxmock.PgxPoolIface)
		subjectID  string
		subjectType string
		roleID     int64
		wantErr    bool
		errMsg     string
	}{
		{
			name: "Happy Path - Delete existing subject role",
			setupMock: func(mockPool pgxmock.PgxPoolIface) {
				mockPool.ExpectExec(`DELETE FROM subject_role WHERE subject_id = \$1 AND subject_type = \$2 AND role_id = \$3`).
					WithArgs("user-123", "user", int64(1)).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
			},
			subjectID:   "user-123",
			subjectType: "user",
			roleID:      1,
			wantErr:     false,
		},
		{
			name: "Error - Subject role not found (0 rows affected)",
			setupMock: func(mockPool pgxmock.PgxPoolIface) {
				mockPool.ExpectExec(`DELETE FROM subject_role WHERE subject_id = \$1 AND subject_type = \$2 AND role_id = \$3`).
					WithArgs("user-123", "user", int64(1)).
					WillReturnResult(pgxmock.NewResult("DELETE", 0))
			},
			subjectID:   "user-123",
			subjectType: "user",
			roleID:      1,
			wantErr:     true,
			errMsg:      "subject role not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			tt.setupMock(mockPool)

			repo := NewSubjectRepository(mockPool)
			err = repo.Delete(context.Background(), tt.subjectID, tt.subjectType, tt.roleID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestAdapter_SubjectRepository_List(t *testing.T) {
	tests := []struct {
		name       string
		pagination *param.PaginationParam
		filter     *param.SubjectListFilterParam
		wantErr    bool
		wantCount  int
		wantTotal  int64
	}{
		{
			name: "Happy Path - List all subject roles",
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
			name: "Happy Path - List with SubjectID filter",
			pagination: &param.PaginationParam{
				Page:  func() *int { i := 1; return &i }(),
				Limit: func() *int { i := 10; return &i }(),
			},
			filter: &param.SubjectListFilterParam{
				SubjectID: func() *string { s := "subject-001"; return &s }(),
			},
			wantErr:   false,
			wantCount: 1,
			wantTotal: 1,
		},
		{
			name: "Happy Path - List with SubjectType filter",
			pagination: &param.PaginationParam{
				Page:  func() *int { i := 1; return &i }(),
				Limit: func() *int { i := 10; return &i }(),
			},
			filter: &param.SubjectListFilterParam{
				SubjectType: func() *string { s := "user"; return &s }(),
			},
			wantErr:   false,
			wantCount: 1,
			wantTotal: 1,
		},
		{
			name: "Happy Path - List with RoleID filter",
			pagination: &param.PaginationParam{
				Page:  func() *int { i := 1; return &i }(),
				Limit: func() *int { i := 10; return &i }(),
			},
			filter: &param.SubjectListFilterParam{
				RoleID: func() *int64 { i := int64(1); return &i }(),
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
		{
			name: "Happy Path - Invalid OrderBy falls back to default",
			pagination: &param.PaginationParam{
				Page:    func() *int { i := 1; return &i }(),
				Limit:   func() *int { i := 10; return &i }(),
				OrderBy: func() *string { s := "created_at"; return &s }(),
			},
			filter:    nil,
			wantErr:   false,
			wantCount: 2,
			wantTotal: 2,
		},
		{
			name: "Happy Path - Valid OrderBy uses specified column",
			pagination: &param.PaginationParam{
				Page:    func() *int { i := 1; return &i }(),
				Limit:   func() *int { i := 10; return &i }(),
				OrderBy: func() *string { s := "subject_id"; return &s }(),
			},
			filter:    nil,
			wantErr:   false,
			wantCount: 2,
			wantTotal: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherRegexp))
			require.NoError(t, err)
			defer mockPool.Close()

			repo := NewSubjectRepository(mockPool)

			// Calculate expected arguments based on filter
			filterArgCount := 0
			if tt.filter != nil {
				if tt.filter.SubjectID != nil {
					filterArgCount += 1
				}
				if tt.filter.SubjectType != nil {
					filterArgCount += 1
				}
				if tt.filter.RoleID != nil {
					filterArgCount += 1
				}
				if tt.filter.RoleUID != nil {
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
			dataRows := pgxmock.NewRows([]string{"subject_id", "subject_type", "role_id", "role_uid", "assigned_at"})
			for i := 0; i < tt.wantCount; i++ {
				dataRows.AddRow(fmt.Sprintf("subject-%d", i+1), "user", int64(1), "role-uid", time.Time{})
			}
			mockPool.ExpectQuery(`SELECT sr\.subject_id, sr\.subject_type, sr\.role_id, r\.uid as role_uid, sr\.assigned_at FROM subject_role sr JOIN role r ON sr\.role_id = r\.id WHERE 1=1`).
				WithArgs(dataAnyArgs...).
				WillReturnRows(dataRows)

			// Set up count expectations
			countRows := pgxmock.NewRows([]string{"count"}).AddRow(tt.wantTotal)
			mockPool.ExpectQuery(`SELECT COUNT\(\*\) FROM subject_role sr JOIN role r ON sr\.role_id = r\.id WHERE 1=1`).
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
