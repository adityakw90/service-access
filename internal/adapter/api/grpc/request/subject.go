package request

import (
	"strings"

	"github.com/adityakw90/service-access-proto/gen/go/subject"
	param "github.com/adityakw90/service-access/internal/core/domain/param"
)

// AssignRoleRequest represents validated subject role assignment request data.
type AssignRoleRequest struct {
	SubjectID   string `validate:"required"`
	SubjectType string `validate:"required"`
	RoleUID     string `validate:"required"`
}

func (r *AssignRoleRequest) ToAssignParams() *param.SubjectAssignRoleParam {
	return &param.SubjectAssignRoleParam{
		SubjectID:   r.SubjectID,
		SubjectType: r.SubjectType,
		RoleUID:     r.RoleUID,
	}
}

// AssignRoleRequestFromPb converts a proto AssignRoleRequest to an AssignRoleRequest.
func AssignRoleRequestFromPb(req *subject.AssignRoleRequest) *AssignRoleRequest {
	return &AssignRoleRequest{
		SubjectID:   strings.TrimSpace(req.GetSubjectId()),
		SubjectType: strings.TrimSpace(req.GetSubjectType()),
		RoleUID:     strings.TrimSpace(req.GetRoleUid()),
	}
}

// RevokeRoleRequest represents validated subject role revocation request data.
type RevokeRoleRequest struct {
	SubjectID   string `validate:"required"`
	SubjectType string `validate:"required"`
	RoleUID     string `validate:"required"`
}

func (r *RevokeRoleRequest) ToRevokeParams() *param.SubjectRevokeRoleParam {
	return &param.SubjectRevokeRoleParam{
		SubjectID:   r.SubjectID,
		SubjectType: r.SubjectType,
		RoleUID:     r.RoleUID,
	}
}

// RevokeRoleRequestFromPb converts a proto RevokeRoleRequest to a RevokeRoleRequest.
func RevokeRoleRequestFromPb(req *subject.RevokeRoleRequest) *RevokeRoleRequest {
	return &RevokeRoleRequest{
		SubjectID:   strings.TrimSpace(req.GetSubjectId()),
		SubjectType: strings.TrimSpace(req.GetSubjectType()),
		RoleUID:     strings.TrimSpace(req.GetRoleUid()),
	}
}

// ToSubjectListFilterParam converts proto ListRequest directly to domain filter param.
func ToSubjectListFilterParam(req *subject.ListRequest) *param.SubjectListFilterParam {
	if req.Filter == nil {
		return &param.SubjectListFilterParam{}
	}

	return &param.SubjectListFilterParam{
		SubjectID:   req.Filter.SubjectId,
		SubjectType: req.Filter.SubjectType,
		RoleUID:     req.Filter.RoleUid,
		Query:       req.Filter.Query,
	}
}
