package repository

import (
	"testing"

	"github.com/adityakw90/service-access/internal/core/domain/param"
	"github.com/stretchr/testify/assert"
)

func TestValidateOrderBy(t *testing.T) {
	// Create a test allowed order by map using RoleOrderBy
	testAllowedOrderBy := map[string]param.RoleOrderBy{
		"id":          param.OrderByRoleID,
		"uid":         param.OrderByRoleUID,
		"group_id":    param.OrderByRoleGroupID,
		"name":        param.OrderByRoleName,
		"description": param.OrderByRoleDescription,
		"created_at":  param.OrderByRoleCreatedAt,
		"updated_at":  param.OrderByRoleUpdatedAt,
	}

	tests := []struct {
		name       string
		pagination *param.PaginationParam
		defaultOrd string
		allowed    map[string]param.RoleOrderBy
		want       string
	}{
		{
			name: "Valid OrderBy - name",
			pagination: &param.PaginationParam{
				OrderBy: func() *string { s := "name"; return &s }(),
			},
			defaultOrd: "created_at",
			allowed:    testAllowedOrderBy,
			want:       "name",
		},
		{
			name: "Valid OrderBy - description",
			pagination: &param.PaginationParam{
				OrderBy: func() *string { s := "description"; return &s }(),
			},
			defaultOrd: "created_at",
			allowed:    testAllowedOrderBy,
			want:       "description",
		},
		{
			name: "Valid OrderBy - id",
			pagination: &param.PaginationParam{
				OrderBy: func() *string { s := "id"; return &s }(),
			},
			defaultOrd: "created_at",
			allowed:    testAllowedOrderBy,
			want:       "id",
		},
		{
			name: "Invalid OrderBy - SQL injection attempt",
			pagination: &param.PaginationParam{
				OrderBy: func() *string { s := "id; DROP TABLE roles; --"; return &s }(),
			},
			defaultOrd: "created_at",
			allowed:    testAllowedOrderBy,
			want:       "created_at", // fallback to default
		},
		{
			name: "Invalid OrderBy - non-existent column",
			pagination: &param.PaginationParam{
				OrderBy: func() *string { s := "nonexistent"; return &s }(),
			},
			defaultOrd: "created_at",
			allowed:    testAllowedOrderBy,
			want:       "created_at", // fallback to default
		},
		{
			name:       "Nil pagination",
			pagination: nil,
			defaultOrd: "created_at",
			allowed:    testAllowedOrderBy,
			want:       "created_at",
		},
		{
			name: "Nil OrderBy",
			pagination: &param.PaginationParam{
				OrderBy: nil,
			},
			defaultOrd: "created_at",
			allowed:    testAllowedOrderBy,
			want:       "created_at",
		},
		{
			name: "Empty OrderBy string",
			pagination: &param.PaginationParam{
				OrderBy: func() *string { s := ""; return &s }(),
			},
			defaultOrd: "created_at",
			allowed:    testAllowedOrderBy,
			want:       "created_at", // empty string is not in map
		},
		{
			name: "Valid OrderBy with different default",
			pagination: &param.PaginationParam{
				OrderBy: func() *string { s := "uid"; return &s }(),
			},
			defaultOrd: "id",
			allowed:    testAllowedOrderBy,
			want:       "uid",
		},
		{
			name: "Invalid OrderBy with custom default",
			pagination: &param.PaginationParam{
				OrderBy: func() *string { s := "invalid"; return &s }(),
			},
			defaultOrd: "updated_at",
			allowed:    testAllowedOrderBy,
			want:       "updated_at",
		},
		{
			name: "Empty allowed map",
			pagination: &param.PaginationParam{
				OrderBy: func() *string { s := "id"; return &s }(),
			},
			defaultOrd: "created_at",
			allowed:    map[string]param.RoleOrderBy{},
			want:       "created_at", // fallback to default since map is empty
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validateOrderBy(tt.pagination, tt.defaultOrd, tt.allowed)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestValidateOrderByWithSubject(t *testing.T) {
	// Test with SubjectOrderBy to ensure generic works with different types
	allowedOrderBySubject := map[string]param.SubjectOrderBy{
		"subject_id":   param.OrderBySubjectID,
		"subject_type": param.OrderBySubjectType,
		"role_id":      param.OrderBySubjectRoleID,
		"assigned_at":  param.OrderBySubjectAssignedAt,
	}

	tests := []struct {
		name       string
		pagination *param.PaginationParam
		defaultOrd string
		want       string
	}{
		{
			name: "Valid Subject OrderBy - subject_type",
			pagination: &param.PaginationParam{
				OrderBy: func() *string { s := "subject_type"; return &s }(),
			},
			defaultOrd: "assigned_at",
			want:       "subject_type",
		},
		{
			name: "Valid Subject OrderBy - role_id",
			pagination: &param.PaginationParam{
				OrderBy: func() *string { s := "role_id"; return &s }(),
			},
			defaultOrd: "assigned_at",
			want:       "role_id",
		},
		{
			name: "Invalid Subject OrderBy",
			pagination: &param.PaginationParam{
				OrderBy: func() *string { s := "invalid_column"; return &s }(),
			},
			defaultOrd: "assigned_at",
			want:       "assigned_at",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validateOrderBy(tt.pagination, tt.defaultOrd, allowedOrderBySubject)
			assert.Equal(t, tt.want, got)
		})
	}
}
