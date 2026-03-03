package response

import (
	"github.com/adityakw90/service-access-proto/gen/go/subject"
	"github.com/adityakw90/service-access/internal/core/domain/model"
)

// ToProtoSubjectRole converts domain SubjectRole to proto SubjectRole.
func ToProtoSubjectRole(sr *model.SubjectRole) *subject.SubjectRole {
	if sr == nil {
		return nil
	}

	return &subject.SubjectRole{
		SubjectId:   sr.SubjectID,
		SubjectType: sr.SubjectType,
		RoleUid:     sr.RoleUID,
		AssignedAt:  toProtoTimestampPB(sr.AssignedAt),
	}
}

// ToProtoSubjectRoleList converts domain SubjectRoles to proto ListResponse.
func ToProtoSubjectRoleList(srs *model.SubjectRoles, meta *model.Meta) *subject.ListResponse {
	items := make([]*subject.SubjectRole, len(srs.Items))
	for i, sr := range srs.Items {
		items[i] = ToProtoSubjectRole(&sr)
	}

	return &subject.ListResponse{
		Items: items,
		Meta:  ToProtoMeta(meta),
	}
}
