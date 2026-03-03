package request_test

import (
	"testing"

	"github.com/adityakw90/service-access/internal/adapter/api/grpc/request"
	"github.com/adityakw90/service-access-proto/gen/go/subject"
)

func TestToSubjectListFilterParam(t *testing.T) {
	tests := []struct {
		name string
		req  *subject.ListRequest
	}{
		{
			name: "Nil filter",
			req:  &subject.ListRequest{},
		},
		{
			name: "With filter",
			req: &subject.ListRequest{
				Filter: &subject.FilterRequest{
					SubjectId:   stringPtr("user-123"),
					SubjectType: stringPtr("user"),
					RoleUid:     stringPtr("role-123"),
					Query:       stringPtr("test"),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := request.ToSubjectListFilterParam(tt.req)
			if got == nil {
				t.Error("ToSubjectListFilterParam() should not return nil")
			}
		})
	}
}

func TestAssignRoleRequestFromPb(t *testing.T) {
	req := request.AssignRoleRequestFromPb(&subject.AssignRoleRequest{
		SubjectId:   "user-123",
		SubjectType: "user",
		RoleUid:     "role-123",
	})
	if req == nil {
		t.Error("AssignRoleRequestFromPb() should not return nil")
	}
}

func TestRevokeRoleRequestFromPb(t *testing.T) {
	req := request.RevokeRoleRequestFromPb(&subject.RevokeRoleRequest{
		SubjectId:   "user-123",
		SubjectType: "user",
		RoleUid:     "role-123",
	})
	if req == nil {
		t.Error("RevokeRoleRequestFromPb() should not return nil")
	}
}

func stringPtr(s string) *string {
	return &s
}
