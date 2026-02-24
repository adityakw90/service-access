package request_test

import (
	"testing"

	"github.com/adityakw90/service-access/internal/adapter/api/grpc/request"
	"github.com/adityakw90/service-access-proto/gen/go/access"
)

func TestToAccessListFilterParam(t *testing.T) {
	subjectID := "user-123"
	subjectType := "user"
	roleUID := "role-123"
	query := "test"

	tests := []struct {
		name string
		req  *access.ListSubjectRolesRequest
	}{
		{
			name: "Nil filter",
			req:  &access.ListSubjectRolesRequest{},
		},
		{
			name: "With filter",
			req: &access.ListSubjectRolesRequest{
				Filter: &access.FilterRequest{
					SubjectId:   &subjectID,
					SubjectType: &subjectType,
					RoleUid:     &roleUID,
					Query:       &query,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := request.ToAccessListFilterParam(tt.req)
			if got == nil {
				t.Error("ToAccessListFilterParam() should not return nil")
			}
		})
	}
}
