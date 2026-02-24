package request

import (
	"strings"

	access "github.com/adityakw90/service-access-proto/gen/go/access"
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
func AssignRoleRequestFromPb(req *access.AssignRoleRequest) *AssignRoleRequest {
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
func RevokeRoleRequestFromPb(req *access.RevokeRoleRequest) *RevokeRoleRequest {
	return &RevokeRoleRequest{
		SubjectID:   strings.TrimSpace(req.GetSubjectId()),
		SubjectType: strings.TrimSpace(req.GetSubjectType()),
		RoleUID:     strings.TrimSpace(req.GetRoleUid()),
	}
}

// ListSubjectRolesRequest represents validated subject role list request data.
type ListSubjectRolesRequest struct {
	Pagination *PaginationRequest
	Filter     *SubjectListFilterRequest
}

// ListSubjectRolesRequestFromPb converts a proto ListSubjectRolesRequest to a ListSubjectRolesRequest.
func ListSubjectRolesRequestFromPb(req *access.ListSubjectRolesRequest) *ListSubjectRolesRequest {
	payload := &ListSubjectRolesRequest{}

	if req.Pagination != nil {
		payload.Pagination = PaginationRequestFromPb(req.GetPagination())
	}

	if req.Filter != nil {
		payload.Filter = subjectListFilterFromPb(req.GetFilter())
	}

	return payload
}

// ToSubjectListParams converts a ListSubjectRolesRequest to domain params.
func (r *ListSubjectRolesRequest) ToSubjectListParams() *param.SubjectListParam {
	var pagination *param.PaginationParam
	if r.Pagination != nil {
		pagination = r.Pagination.ToPaginationParam()
	}

	var filter *param.SubjectListFilterParam
	if r.Filter != nil {
		filter = r.Filter.toSubjectListFilterParam()
	}

	return &param.SubjectListParam{
		Pagination: pagination,
		Filter:     filter,
	}
}

// SubjectListFilterRequest represents validated subject filter request data.
type SubjectListFilterRequest struct {
	SubjectID   *string `validate:"omitempty"`
	SubjectType *string `validate:"omitempty"`
	RoleUID     *string `validate:"omitempty"`
	Query       *string `validate:"omitempty"`
}

// subjectListFilterFromPb converts a proto FilterRequest to a SubjectListFilterRequest.
func subjectListFilterFromPb(req *access.FilterRequest) *SubjectListFilterRequest {
	if req == nil {
		return nil
	}

	r := &SubjectListFilterRequest{}

	if subjectID := strings.TrimSpace(req.GetSubjectId()); subjectID != "" {
		r.SubjectID = &subjectID
	}
	if subjectType := strings.TrimSpace(req.GetSubjectType()); subjectType != "" {
		r.SubjectType = &subjectType
	}
	if roleUID := strings.TrimSpace(req.GetRoleUid()); roleUID != "" {
		r.RoleUID = &roleUID
	}
	if query := strings.TrimSpace(req.GetQuery()); query != "" {
		r.Query = &query
	}

	return r
}

// toSubjectListFilterParam converts a SubjectListFilterRequest to domain params.
func (r *SubjectListFilterRequest) toSubjectListFilterParam() *param.SubjectListFilterParam {
	return &param.SubjectListFilterParam{
		SubjectID:   r.SubjectID,
		SubjectType: r.SubjectType,
		RoleUID:     r.RoleUID,
		Query:       r.Query,
	}
}
