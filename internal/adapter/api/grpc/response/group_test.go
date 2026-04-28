package response_test

import (
	"testing"
	"time"

	"github.com/adityakw90/service-access/internal/adapter/api/grpc/response"
	"github.com/adityakw90/service-access/internal/core/domain/model"
)

func TestToProtoGroup(t *testing.T) {
	tests := []struct {
		name string
		grp  *model.Group
		want string
	}{
		{
			name: "Valid group",
			grp: &model.Group{
				UID:  "group-123",
				Name: "invoice-management",
			},
			want: "group-123",
		},
		{
			name: "Nil group",
			grp:  nil,
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := response.ToProtoGroup(tt.grp)
			if tt.grp == nil && got != nil {
				t.Errorf("ToProtoGroup() should return nil for nil input")
			}
			if got != nil && got.Uid != tt.want {
				t.Errorf("ToProtoGroup() uid = %v, want %v", got.Uid, tt.want)
			}
		})
	}
}

func TestToProtoGroupList(t *testing.T) {
	tests := []struct {
		name   string
		groups *model.Groups
		meta   *model.Meta
	}{
		{
			name: "Valid list",
			groups: &model.Groups{
				Items: []model.Group{
					{UID: "group-1", Name: "admin"},
					{UID: "group-2", Name: "moderator"},
				},
			},
			meta: &model.Meta{Page: 1, Limit: 10, Total: 2, Pages: 1},
		},
		{
			name: "Empty list",
			groups: &model.Groups{
				Items: []model.Group{},
			},
			meta: &model.Meta{Page: 1, Limit: 10, Total: 0, Pages: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := response.ToProtoGroupList(tt.groups, tt.meta)
			if len(got.Items) != len(tt.groups.Items) {
				t.Errorf("ToProtoGroupList() items count = %v, want %v", len(got.Items), len(tt.groups.Items))
			}
			if got.Meta == nil {
				t.Error("ToProtoGroupList() meta should not be nil")
			}
		})
	}
}

func TestToProtoGroupPermission(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name string
		gp   *model.GroupPermission
		want string
	}{
		{
			name: "Valid group permission",
			gp: &model.GroupPermission{
				UID:                   "gperm-123",
				GroupUID:              "group-123",
				PermissionUID:         "perm-123",
				PermissionResource:    "invoices",
				PermissionAction:      "read",
				PermissionDescription: "Read invoices",
				CreatedAt:             now,
			},
			want: "gperm-123",
		},
		{
			name: "Nil group permission",
			gp:   nil,
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := response.ToProtoGroupPermission(tt.gp)
			if tt.gp == nil && got != nil {
				t.Errorf("ToProtoGroupPermission() should return nil for nil input")
			}
			if got != nil && got.Uid != tt.want {
				t.Errorf("ToProtoGroupPermission() uid = %v, want %v", got.Uid, tt.want)
			}
		})
	}
}

func TestToProtoGroupPermissionList(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name string
		gps  *model.GroupPermissions
		meta *model.Meta
	}{
		{
			name: "Valid list",
			gps: &model.GroupPermissions{
				Items: []model.GroupPermission{
					{UID: "gperm-1", PermissionUID: "perm-1", CreatedAt: now},
					{UID: "gperm-2", PermissionUID: "perm-2", CreatedAt: now},
				},
			},
			meta: &model.Meta{Page: 1, Limit: 10, Total: 2, Pages: 1},
		},
		{
			name: "Empty list",
			gps: &model.GroupPermissions{
				Items: []model.GroupPermission{},
			},
			meta: &model.Meta{Page: 1, Limit: 10, Total: 0, Pages: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := response.ToProtoGroupPermissionList(tt.gps, tt.meta)
			if len(got.Items) != len(tt.gps.Items) {
				t.Errorf("ToProtoGroupPermissionList() items count = %v, want %v", len(got.Items), len(tt.gps.Items))
			}
			if got.Meta == nil {
				t.Error("ToProtoGroupPermissionList() meta should not be nil")
			}
		})
	}
}
