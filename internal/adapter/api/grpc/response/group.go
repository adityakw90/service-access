package response

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/adityakw90/service-access-proto/gen/go/group"
	"github.com/adityakw90/service-access/internal/core/domain/model"
)

func ToGroupResponse(g *model.Group) *group.Group {
	return &group.Group{
		Uid:         g.UID,
		Name:        g.Name,
		Description: g.Description,
		CreatedAt:   timestamppb.New(g.CreatedAt),
		UpdatedAt:   timestamppb.New(g.UpdatedAt),
	}
}

// ToProtoGroup converts domain Group to proto Group.
func ToProtoGroup(g *model.Group) *group.Group {
	if g == nil {
		return nil
	}

	return &group.Group{
		Uid:         g.UID,
		Name:        g.Name,
		Description: g.Description,
		CreatedAt:   toProtoTimestampPB(g.CreatedAt),
		UpdatedAt:   toProtoTimestampPB(g.UpdatedAt),
	}
}

// ToProtoGroupList converts domain Groups to proto ListResponse.
func ToProtoGroupList(groups *model.Groups, meta *model.Meta) *group.ListResponse {
	items := make([]*group.Group, len(groups.Items))
	for i, g := range groups.Items {
		items[i] = ToProtoGroup(&g)
	}

	return &group.ListResponse{
		Items: items,
		Meta:  ToProtoMeta(meta),
	}
}

// ToProtoGroupPermission converts domain GroupPermission to proto GroupPermission.
func ToProtoGroupPermission(gp *model.GroupPermission) *group.GroupPermission {
	if gp == nil {
		return nil
	}

	return &group.GroupPermission{
		Uid:                 gp.UID,
		GroupUid:            gp.GroupUID,
		PermissionUid:       gp.PermissionUID,
		PermissionResource:  gp.PermissionResource,
		PermissionAction:    gp.PermissionAction,
		PermissionDescription: gp.PermissionDescription,
		CreatedAt:           toProtoTimestampPB(gp.CreatedAt),
	}
}

// ToProtoGroupPermissionList converts domain GroupPermissions to proto ListPermissionsResponse.
func ToProtoGroupPermissionList(gps *model.GroupPermissions, meta *model.Meta) *group.ListPermissionsResponse {
	items := make([]*group.GroupPermission, len(gps.Items))
	for i, gp := range gps.Items {
		items[i] = ToProtoGroupPermission(&gp)
	}

	return &group.ListPermissionsResponse{
		Items: items,
		Meta:  ToProtoMeta(meta),
	}
}
