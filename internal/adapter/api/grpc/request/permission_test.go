package request_test

import (
	"testing"

	"github.com/adityakw90/service-access/internal/adapter/api/grpc/request"
	"github.com/adityakw90/service-access-proto/gen/go/common"
	"github.com/adityakw90/service-access-proto/gen/go/permission"
)

func TestToPermissionCreateParam(t *testing.T) {
	tests := []struct {
		name    string
		req     *permission.CreateRequest
		want    string // resource for verification
		wantErr bool
	}{
		{
			name: "Valid request",
			req: &permission.CreateRequest{
				Resource:    "invoices",
				Action:      "read",
				Description: "Read invoices",
			},
			want:    "invoices",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := request.ToPermissionCreateParam(tt.req)
			if got.Resource != tt.want {
				t.Errorf("ToPermissionCreateParam() resource = %v, want %v", got.Resource, tt.want)
			}
		})
	}
}

func TestToPermissionUpdateParam(t *testing.T) {
	tests := []struct {
		name    string
		req     *permission.UpdateRequest
		want    string // resource for verification
	}{
		{
			name: "Valid request",
			req: &permission.UpdateRequest{
				Uid:         "perm-123",
				Resource:    "invoices",
				Action:      "write",
				Description: "Write invoices",
			},
			want: "invoices",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := request.ToPermissionUpdateParam(tt.req)
			if got.Resource == nil || *got.Resource != tt.want {
				t.Errorf("ToPermissionUpdateParam() resource = %v, want %v", got.Resource, tt.want)
			}
		})
	}
}

func TestToPermissionListFilterParam(t *testing.T) {
	tests := []struct {
		name string
		req  *permission.ListRequest
		want *struct{ uids []string }
	}{
		{
			name: "With UIDs",
			req: &permission.ListRequest{
				Filter: &permission.FilterRequest{
					Uids: []string{"uid1", "uid2"},
				},
			},
			want: &struct{ uids []string }{uids: []string{"uid1", "uid2"}},
		},
		{
			name: "Nil filter",
			req:  &permission.ListRequest{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := request.ToPermissionListFilterParam(tt.req)
			if tt.want != nil && tt.want.uids != nil {
				if len(got.UIDs) != len(tt.want.uids) {
					t.Errorf("ToPermissionListFilterParam() uids length = %v, want %v", len(got.UIDs), len(tt.want.uids))
				}
			}
		})
	}
}

func TestToPaginationParam(t *testing.T) {
	tests := []struct {
		name string
		req  *permission.ListRequest
	}{
		{
			name: "With pagination",
			req: &permission.ListRequest{
				Pagination: &common.Pagination{
					Page:  1,
					Limit: 10,
				},
			},
		},
		{
			name: "Nil pagination",
			req:  &permission.ListRequest{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := request.ToPaginationParam(tt.req.Pagination)
			if got == nil {
				t.Error("ToPaginationParam() should not return nil")
			}
		})
	}
}
