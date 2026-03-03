package observer

import (
	"github.com/adityakw90/go-monitoring"
	"github.com/adityakw90/service-access/internal/core/domain/signal"
	"go.opentelemetry.io/otel/attribute"
)

func NewRoleObserver(
	logger monitoring.Logger,
	tracer monitoring.Tracer,
) *serviceObserver[signal.SignalRole] {
	return NewServiceObserver(
		logger,
		tracer,
		func(s signal.SignalRole) []attribute.KeyValue {
			attrs := []attribute.KeyValue{
				attribute.String("operation", s.Operation),
			}
			if s.UID != nil {
				attrs = append(attrs, attribute.String("role.uid", *s.UID))
			}
			if s.GroupUID != nil {
				attrs = append(attrs, attribute.String("role.group_uid", *s.GroupUID))
			}
			if s.Name != nil {
				attrs = append(attrs, attribute.String("role.name", *s.Name))
			}
			if s.Description != nil {
				attrs = append(attrs, attribute.String("role.description", *s.Description))
			}
			if s.CreatedAt != nil {
				attrs = append(attrs, attribute.Int64("role.created_at", s.CreatedAt.Unix()))
			}
			if s.UpdatedAt != nil {
				attrs = append(attrs, attribute.Int64("role.updated_at", s.UpdatedAt.Unix()))
			}
			return attrs
		},
		func(s signal.SignalRole) map[string]any {
			fields := map[string]any{
				"operation": s.Operation,
			}
			if s.UID != nil {
				fields["role.uid"] = *s.UID
			}
			if s.GroupUID != nil {
				fields["role.group_uid"] = *s.GroupUID
			}
			if s.Name != nil {
				fields["role.name"] = *s.Name
			}
			if s.Description != nil {
				fields["role.description"] = *s.Description
			}
			if s.CreatedAt != nil {
				fields["role.created_at"] = *s.CreatedAt
			}
			if s.UpdatedAt != nil {
				fields["role.updated_at"] = *s.UpdatedAt
			}
			return fields
		},
	)
}

func NewRolePermissionObserver(
	logger monitoring.Logger,
	tracer monitoring.Tracer,
) *serviceObserver[signal.SignalRolePermission] {
	return NewServiceObserver(
		logger,
		tracer,
		func(s signal.SignalRolePermission) []attribute.KeyValue {
			attrs := []attribute.KeyValue{
				attribute.String("operation", s.Operation),
			}
			if s.RoleUID != nil {
				attrs = append(attrs, attribute.String("role_permission.group_uid", *s.RoleUID))
			}
			if s.GroupPermissionUID != nil {
				attrs = append(attrs, attribute.String("role_permission.group_permission_uid", *s.GroupPermissionUID))
			}
			if s.PermissionUID != nil {
				attrs = append(attrs, attribute.String("role_permission.permission_uid", *s.PermissionUID))
			}
			if s.PermissionResource != nil {
				attrs = append(attrs, attribute.String("role_permission.permission_resource", *s.PermissionResource))
			}
			if s.PermissionAction != nil {
				attrs = append(attrs, attribute.String("role_permission.permission_action", *s.PermissionAction))
			}
			if s.PermissionDescription != nil {
				attrs = append(attrs, attribute.String("role_permission.permission_description", *s.PermissionDescription))
			}
			if s.CreatedAt != nil {
				attrs = append(attrs, attribute.Int64("role_permission.created_at", s.CreatedAt.Unix()))
			}
			if len(s.GroupPermissionUIDs) > 0 {
				attrs = append(attrs, attribute.StringSlice("role_permission.group_permission_uids", s.GroupPermissionUIDs))
			}
			if len(s.PermissionUIDs) > 0 {
				attrs = append(attrs, attribute.StringSlice("role_permission.permission_uids", s.PermissionUIDs))
			}
			return attrs
		},
		func(s signal.SignalRolePermission) map[string]any {
			fields := map[string]any{
				"operation": s.Operation,
			}
			if s.RoleUID != nil {
				fields["role_permission.group_uid"] = *s.RoleUID
			}
			if s.GroupPermissionUID != nil {
				fields["role_permission.group_permission_uid"] = *s.GroupPermissionUID
			}
			if s.PermissionUID != nil {
				fields["role_permission.permission_uid"] = *s.PermissionUID
			}
			if s.PermissionResource != nil {
				fields["role_permission.permission_resource"] = *s.PermissionResource
			}
			if s.PermissionAction != nil {
				fields["role_permission.permission_action"] = *s.PermissionAction
			}
			if s.PermissionDescription != nil {
				fields["role_permission.permission_description"] = *s.PermissionDescription
			}
			if s.CreatedAt != nil {
				fields["role_permission.created_at"] = *s.CreatedAt
			}
			if len(s.GroupPermissionUIDs) > 0 {
				fields["role_permission.group_permission_uids"] = s.GroupPermissionUIDs
			}
			if len(s.PermissionUIDs) > 0 {
				fields["role_permission.permission_uids"] = s.PermissionUIDs
			}
			return fields
		},
	)
}
