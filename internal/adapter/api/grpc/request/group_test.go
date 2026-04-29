package request_test

import (
	"testing"

	"github.com/adityakw90/service-access-proto/gen/go/group"
	"github.com/adityakw90/service-access/internal/adapter/api/grpc/request"
)

func TestToGroupCreateParam(t *testing.T) {
	tests := []struct {
		name string
		req  *group.CreateRequest
		want string
	}{
		{
			name: "Valid request",
			req: &group.CreateRequest{
				Name:        "invoice-management",
				Description: "Invoice permissions",
			},
			want: "invoice-management",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := request.ToGroupCreateParam(tt.req)
			if got.Name != tt.want {
				t.Errorf("ToGroupCreateParam() name = %v, want %v", got.Name, tt.want)
			}
		})
	}
}

func TestToGroupUpdateParam(t *testing.T) {
	tests := []struct {
		name string
		req  *group.UpdateRequest
	}{
		{
			name: "Valid request with all fields",
			req: &group.UpdateRequest{
				Name:        "updated-name",
				Description: "Updated description",
			},
		},
		{
			name: "Valid request with partial fields",
			req: &group.UpdateRequest{
				Name: "name-only",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := request.ToGroupUpdateParam(tt.req)
			if got.Name == nil && got.Description == nil {
				t.Error("ToGroupUpdateParam() should have at least one field set")
			}
		})
	}
}

func TestToGroupListFilterParam(t *testing.T) {
	name := "admin"
	query := "test"

	tests := []struct {
		name string
		req  *group.ListRequest
	}{
		{
			name: "Nil filter",
			req:  &group.ListRequest{},
		},
		{
			name: "With filter",
			req: &group.ListRequest{
				Filter: &group.FilterRequest{
					Uids:  []string{"uid1"},
					Name:  &name,
					Query: &query,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := request.ToGroupListFilterParam(tt.req)
			if got == nil {
				t.Error("ToGroupListFilterParam() should not return nil")
			}
		})
	}
}

func TestToGroupPermissionListFilterParam(t *testing.T) {
	resource := "invoices"
	action := "read"
	query := "test"

	tests := []struct {
		name string
		req  *group.ListPermissionsRequest
	}{
		{
			name: "Nil filter",
			req:  &group.ListPermissionsRequest{},
		},
		{
			name: "With filter",
			req: &group.ListPermissionsRequest{
				Filter: &group.FilterPermissionRequest{
					Uids:           []string{"uid1"},
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
			got := request.ToGroupPermissionListFilterParam(tt.req)
			if got == nil {
				t.Error("ToGroupPermissionListFilterParam() should not return nil")
			}
		})
	}
}
