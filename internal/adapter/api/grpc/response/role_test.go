package response_test

import (
	"testing"
	"time"

	"github.com/adityakw90/service-access/internal/adapter/api/grpc/response"
	"github.com/adityakw90/service-access/internal/core/domain/model"
)

func TestToProtoRole(t *testing.T) {
	tests := []struct {
		name string
		role *model.Role
		want string
	}{
		{
			name: "Valid role",
			role: &model.Role{
				UID:      "role-123",
				GroupUID: "group-123",
				Name:     "admin",
			},
			want: "role-123",
		},
		{
			name: "Nil role",
			role: nil,
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := response.ToProtoRole(tt.role)
			if tt.role == nil && got != nil {
				t.Errorf("ToProtoRole() should return nil for nil input")
			}
			if got != nil && got.Uid != tt.want {
				t.Errorf("ToProtoRole() uid = %v, want %v", got.Uid, tt.want)
			}
		})
	}
}

func TestToProtoRoleList(t *testing.T) {
	tests := []struct {
		name  string
		roles *model.Roles
		meta  *model.Meta
	}{
		{
			name: "Valid list",
			roles: &model.Roles{
				Items: []model.Role{
					{UID: "role-1", GroupUID: "group-1", Name: "admin"},
					{UID: "role-2", GroupUID: "group-1", Name: "moderator"},
				},
			},
			meta: &model.Meta{Page: 1, Limit: 10, Total: 2, Pages: 1},
		},
		{
			name: "Empty list",
			roles: &model.Roles{
				Items: []model.Role{},
			},
			meta: &model.Meta{Page: 1, Limit: 10, Total: 0, Pages: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := response.ToProtoRoleList(tt.roles, tt.meta)
			if len(got.Items) != len(tt.roles.Items) {
				t.Errorf("ToProtoRoleList() items count = %v, want %v", len(got.Items), len(tt.roles.Items))
			}
			if got.Meta == nil {
				t.Error("ToProtoRoleList() meta should not be nil")
			}
		})
	}
}

func TestToProtoRolePermission(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name string
		rp   *model.RolePermission
		want string
	}{
		{
			name: "Valid role permission",
			rp: &model.RolePermission{
				RoleUID:               "role-123",
				GroupPermissionUID:    "gperm-123",
				PermissionUID:         "perm-123",
				PermissionResource:    "invoices",
				PermissionAction:      "read",
				PermissionDescription: "Read invoices",
				CreatedAt:             now,
			},
			want: "role-123",
		},
		{
			name: "Nil role permission",
			rp:   nil,
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := response.ToProtoRolePermission(tt.rp)
			if tt.rp == nil && got != nil {
				t.Errorf("ToProtoRolePermission() should return nil for nil input")
			}
			if got != nil && got.RoleUid != tt.want {
				t.Errorf("ToProtoRolePermission() roleUid = %v, want %v", got.RoleUid, tt.want)
			}
		})
	}
}

func TestToProtoRolePermissionList(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name string
		rps  *model.RolePermissions
		meta *model.Meta
	}{
		{
			name: "Valid list",
			rps: &model.RolePermissions{
				Items: []model.RolePermission{
					{RoleUID: "role-1", PermissionUID: "perm-1", CreatedAt: now},
					{RoleUID: "role-1", PermissionUID: "perm-2", CreatedAt: now},
				},
			},
			meta: &model.Meta{Page: 1, Limit: 10, Total: 2, Pages: 1},
		},
		{
			name: "Empty list",
			rps: &model.RolePermissions{
				Items: []model.RolePermission{},
			},
			meta: &model.Meta{Page: 1, Limit: 10, Total: 0, Pages: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := response.ToProtoRolePermissionList(tt.rps, tt.meta)
			if len(got.Items) != len(tt.rps.Items) {
				t.Errorf("ToProtoRolePermissionList() items count = %v, want %v", len(got.Items), len(tt.rps.Items))
			}
			if got.Meta == nil {
				t.Error("ToProtoRolePermissionList() meta should not be nil")
			}
		})
	}
}
