package observer

import (
	"github.com/adityakw90/go-monitoring"
	"github.com/adityakw90/service-access/internal/core/domain/signal"
	"go.opentelemetry.io/otel/attribute"
)

func NewGroupObserver(
	logger monitoring.Logger,
	tracer monitoring.Tracer,
) *serviceObserver[signal.SignalGroup] {
	return NewServiceObserver(
		logger,
		tracer,
		func(s signal.SignalGroup) []attribute.KeyValue {
			attrs := []attribute.KeyValue{
				attribute.String("operation", s.Operation),
			}
			if s.UID != nil {
				attrs = append(attrs, attribute.String("group.uid", *s.UID))
			}
			if s.Name != nil {
				attrs = append(attrs, attribute.String("group.name", *s.Name))
			}
			if s.Description != nil {
				attrs = append(attrs, attribute.String("group.description", *s.Description))
			}
			if s.CreatedAt != nil {
				attrs = append(attrs, attribute.Int64("group.created_at", s.CreatedAt.Unix()))
			}
			if s.UpdatedAt != nil {
				attrs = append(attrs, attribute.Int64("group.updated_at", s.UpdatedAt.Unix()))
			}
			return attrs
		},
		func(s signal.SignalGroup) map[string]any {
			fields := map[string]any{
				"operation": s.Operation,
			}
			if s.UID != nil {
				fields["group.uid"] = *s.UID
			}
			if s.Name != nil {
				fields["group.name"] = *s.Name
			}
			if s.Description != nil {
				fields["group.description"] = *s.Description
			}
			if s.CreatedAt != nil {
				fields["group.created_at"] = *s.CreatedAt
			}
			if s.UpdatedAt != nil {
				fields["group.updated_at"] = *s.UpdatedAt
			}
			return fields
		},
	)
}

func NewGroupPermissionObserver(
	logger monitoring.Logger,
	tracer monitoring.Tracer,
) *serviceObserver[signal.SignalGroupPermission] {
	return NewServiceObserver(
		logger,
		tracer,
		func(s signal.SignalGroupPermission) []attribute.KeyValue {
			attrs := []attribute.KeyValue{
				attribute.String("operation", s.Operation),
			}
			if s.UID != nil {
				attrs = append(attrs, attribute.String("group_permission.uid", *s.UID))
			}
			if s.GroupUID != nil {
				attrs = append(attrs, attribute.String("group_permission.group_uid", *s.GroupUID))
			}
			if s.PermissionUID != nil {
				attrs = append(attrs, attribute.String("group_permission.permission_uid", *s.PermissionUID))
			}
			if s.PermissionResource != nil {
				attrs = append(attrs, attribute.String("group_permission.permission_resource", *s.PermissionResource))
			}
			if s.PermissionAction != nil {
				attrs = append(attrs, attribute.String("group_permission.permission_action", *s.PermissionAction))
			}
			if s.PermissionDescription != nil {
				attrs = append(attrs, attribute.String("group_permission.permission_description", *s.PermissionDescription))
			}
			if s.CreatedAt != nil {
				attrs = append(attrs, attribute.Int64("group_permission.created_at", s.CreatedAt.Unix()))
			}
			if len(s.PermissionUIDs) > 0 {
				attrs = append(attrs, attribute.StringSlice("group_permission.permission_uids", s.PermissionUIDs))
			}
			return attrs
		},
		func(s signal.SignalGroupPermission) map[string]any {
			fields := map[string]any{
				"operation": s.Operation,
			}
			if s.UID != nil {
				fields["group_permission.uid"] = *s.UID
			}
			if s.GroupUID != nil {
				fields["group_permission.group_uid"] = *s.GroupUID
			}
			if s.PermissionUID != nil {
				fields["group_permission.permission_uid"] = *s.PermissionUID
			}
			if s.PermissionResource != nil {
				fields["group_permission.permission_resource"] = *s.PermissionResource
			}
			if s.PermissionAction != nil {
				fields["group_permission.permission_action"] = *s.PermissionAction
			}
			if s.PermissionDescription != nil {
				fields["group_permission.permission_description"] = *s.PermissionDescription
			}
			if s.CreatedAt != nil {
				fields["group_permission.created_at"] = *s.CreatedAt
			}
			if len(s.PermissionUIDs) > 0 {
				fields["group_permission.permission_uids"] = s.PermissionUIDs
			}
			return fields
		},
	)
}
