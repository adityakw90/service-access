package request_test

import (
	"testing"

	"github.com/adityakw90/service-access/internal/adapter/api/grpc/request"
	"github.com/adityakw90/service-access-proto/gen/go/role"
)

func TestToRoleCreateParam(t *testing.T) {
	tests := []struct {
		name    string
		req     *role.CreateRequest
		want    string
		wantErr bool
	}{
		{
			name: "Valid request",
			req: &role.CreateRequest{
				GroupUid:    "group-123",
				Name:        "admin",
				Description: "Administrator role",
			},
			want:    "group-123",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := request.ToRoleCreateParam(tt.req)
			if got.GroupUID != tt.want {
				t.Errorf("ToRoleCreateParam() groupUID = %v, want %v", got.GroupUID, tt.want)
			}
		})
	}
}

func TestToRoleUpdateParam(t *testing.T) {
	tests := []struct {
		name string
		req  *role.UpdateRequest
	}{
		{
			name: "Valid request with all fields",
			req: &role.UpdateRequest{
				Name:        "super-admin",
				Description: "Super administrator",
			},
		},
		{
			name: "Valid request with partial fields",
			req: &role.UpdateRequest{
				Name: "moderator",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := request.ToRoleUpdateParam(tt.req)
			if got.Name == nil && got.Description == nil {
				t.Error("ToRoleUpdateParam() should have at least one field set")
			}
		})
	}
}

func TestToRoleListFilterParam(t *testing.T) {
	name := "admin"
	query := "test"

	tests := []struct {
		name string
		req  *role.ListRequest
	}{
		{
			name: "Nil filter",
			req:  &role.ListRequest{},
		},
		{
			name: "With filter",
			req: &role.ListRequest{
				Filter: &role.FilterRequest{
					Uids:      []string{"uid1"},
					GroupUids: []string{"group-1"},
					Name:      &name,
					Query:     &query,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := request.ToRoleListFilterParam(tt.req)
			if got == nil {
				t.Error("ToRoleListFilterParam() should not return nil")
			}
		})
	}
}

func TestToRolePermissionListFilterParam(t *testing.T) {
	resource := "invoices"
	action := "read"
	query := "test"

	tests := []struct {
		name string
		req  *role.ListPermissionsRequest
	}{
		{
			name: "Nil filter",
			req:  &role.ListPermissionsRequest{},
		},
		{
			name: "With filter",
			req: &role.ListPermissionsRequest{
				Filter: &role.FilterPermissionRequest{
					PermissionUids: []string{"perm-1"},
					Resource:       &resource,
					Action:         &action,
					Query:          &query,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := request.ToRolePermissionListFilterParam(tt.req)
			if got == nil {
				t.Error("ToRolePermissionListFilterParam() should not return nil")
			}
		})
	}
}
