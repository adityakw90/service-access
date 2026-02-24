package request

import (
	"strings"

	access "github.com/adityakw90/service-access-proto/gen/go/access"
	param "github.com/adityakw90/service-access/internal/core/domain/param"
)

// AccessRequest represents validated access request data.
type AccessRequest struct {
	SubjectID   string `validate:"required"`
	SubjectType string `validate:"required"`
	Resource    string `validate:"required"`
	Action      string `validate:"required"`
}

func (r *AccessRequest) ToAccessParams() *param.AccessParams {
	return &param.AccessParams{
		SubjectID:   r.SubjectID,
		SubjectType: r.SubjectType,
		Resource:    r.Resource,
		Action:      r.Action,
	}
}

func AccessRequestFromPb(req *access.CheckAccessRequest) *AccessRequest {
	payload := &AccessRequest{
		SubjectID:   strings.TrimSpace(req.GetSubjectId()),
		SubjectType: strings.TrimSpace(req.GetSubjectType()),
		Resource:    strings.TrimSpace(req.GetResource()),
		Action:      strings.TrimSpace(req.GetAction()),
	}

	return payload
}

// ToAccessListFilterParam converts proto ListSubjectRolesRequest directly to domain filter param.
func ToAccessListFilterParam(req *access.ListSubjectRolesRequest) *param.SubjectListFilterParam {
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
