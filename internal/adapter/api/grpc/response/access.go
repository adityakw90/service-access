package response

import (
	"github.com/adityakw90/service-access-proto/gen/go/access"
	"github.com/adityakw90/service-access/internal/core/domain/model"
)

func ToAccessResponse(allowed bool, reason string) *access.CheckAccessResponse {
	return &access.CheckAccessResponse{
		Allowed: allowed,
		Reason:  reason,
	}
}

// ToProtoSubjectRole converts domain SubjectRole to proto SubjectRole.
func ToProtoSubjectRole(sr *model.SubjectRole) *access.SubjectRole {
	if sr == nil {
		return nil
	}

	return &access.SubjectRole{
		SubjectId:   sr.SubjectID,
		SubjectType: sr.SubjectType,
		RoleUid:     sr.RoleUID,
		AssignedAt:  toProtoTimestampPB(sr.AssignedAt),
	}
}

// ToProtoSubjectRoleList converts domain SubjectRoles to proto ListSubjectRolesResponse.
func ToProtoSubjectRoleList(srs *model.SubjectRoles, meta *model.Meta) *access.ListSubjectRolesResponse {
	items := make([]*access.SubjectRole, len(srs.Items))
	for i, sr := range srs.Items {
		items[i] = ToProtoSubjectRole(&sr)
	}

	return &access.ListSubjectRolesResponse{
		Items: items,
		Meta:  ToProtoMeta(meta),
	}
}
