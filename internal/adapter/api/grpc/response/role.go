package response

import (
	"github.com/adityakw90/service-access-proto/gen/go/role"
	"github.com/adityakw90/service-access/internal/core/domain/model"
)

// ToProtoRole converts domain Role to proto Role.
func ToProtoRole(r *model.Role) *role.Role {
	if r == nil {
		return nil
	}

	return &role.Role{
		Uid:         r.UID,
		GroupUid:    r.GroupUID,
		Name:        r.Name,
		Description: r.Description,
		CreatedAt:   toProtoTimestampPB(r.CreatedAt),
		UpdatedAt:   toProtoTimestampPB(r.UpdatedAt),
	}
}

// ToProtoRoleList converts domain Roles to proto ListResponse.
func ToProtoRoleList(roles *model.Roles, meta *model.Meta) *role.ListResponse {
	items := make([]*role.Role, len(roles.Items))
	for i, r := range roles.Items {
		items[i] = ToProtoRole(&r)
	}

	return &role.ListResponse{
		Items: items,
		Meta:  ToProtoMeta(meta),
	}
}

// ToProtoRolePermission converts domain RolePermission to proto RolePermission.
func ToProtoRolePermission(rp *model.RolePermission) *role.RolePermission {
	if rp == nil {
		return nil
	}

	return &role.RolePermission{
		RoleUid:               rp.RoleUID,
		GroupPermissionUid:    rp.GroupPermissionUID,
		PermissionUid:         rp.PermissionUID,
		PermissionResource:    rp.PermissionResource,
		PermissionAction:      rp.PermissionAction,
		PermissionDescription: rp.PermissionDescription,
		CreatedAt:             toProtoTimestampPB(rp.CreatedAt),
	}
}

// ToProtoRolePermissionList converts domain RolePermissions to proto ListPermissionsResponse.
func ToProtoRolePermissionList(rps *model.RolePermissions, meta *model.Meta) *role.ListPermissionsResponse {
	items := make([]*role.RolePermission, len(rps.Items))
	for i, rp := range rps.Items {
		items[i] = ToProtoRolePermission(&rp)
	}

	return &role.ListPermissionsResponse{
		Items: items,
		Meta:  ToProtoMeta(meta),
	}
}
