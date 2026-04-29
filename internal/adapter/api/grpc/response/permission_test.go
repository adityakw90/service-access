package response_test

import (
	"testing"
	"time"

	"github.com/adityakw90/service-access/internal/adapter/api/grpc/response"
	"github.com/adityakw90/service-access/internal/core/domain/model"
)

func TestToProtoPermission(t *testing.T) {
	tests := []struct {
		name string
		perm *model.Permission
		want string
	}{
		{
			name: "Valid permission",
			perm: &model.Permission{
				UID:         "perm-123",
				Resource:    "invoices",
				Action:      "read",
				Description: "Read invoices",
			},
			want: "perm-123",
		},
		{
			name: "Nil permission",
			perm: nil,
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := response.ToProtoPermission(tt.perm)
			if tt.perm == nil && got != nil {
				t.Errorf("ToProtoPermission() should return nil for nil input")
			}
			if got != nil && got.Uid != tt.want {
				t.Errorf("ToProtoPermission() uid = %v, want %v", got.Uid, tt.want)
			}
		})
	}
}

func TestToProtoPermissionList(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name  string
		perms *model.Permissions
		meta  *model.Meta
	}{
		{
			name: "Valid list",
			perms: &model.Permissions{
				Items: []model.Permission{
					{UID: "perm-1", Resource: "invoices", Action: "read", CreatedAt: now, UpdatedAt: now},
					{UID: "perm-2", Resource: "users", Action: "write", CreatedAt: now, UpdatedAt: now},
				},
			},
			meta: &model.Meta{Page: 1, Limit: 10, Total: 2, Pages: 1},
		},
		{
			name: "Empty list",
			perms: &model.Permissions{
				Items: []model.Permission{},
			},
			meta: &model.Meta{Page: 1, Limit: 10, Total: 0, Pages: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := response.ToProtoPermissionList(tt.perms, tt.meta)
			if len(got.Items) != len(tt.perms.Items) {
				t.Errorf("ToProtoPermissionList() items count = %v, want %v", len(got.Items), len(tt.perms.Items))
			}
			if got.Meta == nil {
				t.Error("ToProtoPermissionList() meta should not be nil")
			}
		})
	}
}
