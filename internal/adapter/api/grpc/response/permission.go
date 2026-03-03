package response

import (
	"github.com/adityakw90/service-access-proto/gen/go/permission"
	"github.com/adityakw90/service-access/internal/core/domain/model"
)

// ToProtoPermission converts domain Permission to proto Permission.
func ToProtoPermission(p *model.Permission) *permission.Permission {
	if p == nil {
		return nil
	}

	return &permission.Permission{
		Uid:         p.UID,
		Resource:    p.Resource,
		Action:      p.Action,
		Description: p.Description,
		CreatedAt:   toProtoTimestampPB(p.CreatedAt),
		UpdatedAt:   toProtoTimestampPB(p.UpdatedAt),
	}
}

// ToProtoPermissionList converts domain Permissions to proto ListResponse.
func ToProtoPermissionList(perms *model.Permissions, meta *model.Meta) *permission.ListResponse {
	items := make([]*permission.Permission, len(perms.Items))
	for i, p := range perms.Items {
		items[i] = ToProtoPermission(&p)
	}

	return &permission.ListResponse{
		Items: items,
		Meta:  ToProtoMeta(meta),
	}
}
