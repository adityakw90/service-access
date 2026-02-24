package response_test

import (
	"testing"
	"time"

	"github.com/adityakw90/service-access/internal/adapter/api/grpc/response"
	"github.com/adityakw90/service-access/internal/core/domain/model"
)

func TestToProtoSubjectRole(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name string
		sr   *model.SubjectRole
		want string
	}{
		{
			name: "Valid subject role",
			sr: &model.SubjectRole{
				SubjectID:   "user-123",
				SubjectType: "user",
				RoleUID:     "role-123",
				AssignedAt:  now,
			},
			want: "user-123",
		},
		{
			name: "Nil subject role",
			sr:   nil,
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := response.ToProtoSubjectRole(tt.sr)
			if tt.sr == nil && got != nil {
				t.Errorf("ToProtoSubjectRole() should return nil for nil input")
			}
			if got != nil && got.SubjectId != tt.want {
				t.Errorf("ToProtoSubjectRole() subjectId = %v, want %v", got.SubjectId, tt.want)
			}
		})
	}
}

func TestToProtoSubjectRoleList(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name string
		srs  *model.SubjectRoles
		meta *model.Meta
	}{
		{
			name: "Valid list",
			srs: &model.SubjectRoles{
				Items: []model.SubjectRole{
					{SubjectID: "user-123", RoleUID: "role-1", AssignedAt: now},
					{SubjectID: "user-123", RoleUID: "role-2", AssignedAt: now},
				},
			},
			meta: &model.Meta{Page: 1, Limit: 10, Total: 2, Pages: 1},
		},
		{
			name: "Empty list",
			srs: &model.SubjectRoles{
				Items: []model.SubjectRole{},
			},
			meta: &model.Meta{Page: 1, Limit: 10, Total: 0, Pages: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := response.ToProtoSubjectRoleList(tt.srs, tt.meta)
			if len(got.Items) != len(tt.srs.Items) {
				t.Errorf("ToProtoSubjectRoleList() items count = %v, want %v", len(got.Items), len(tt.srs.Items))
			}
			if got.Meta == nil {
				t.Error("ToProtoSubjectRoleList() meta should not be nil")
			}
		})
	}
}
