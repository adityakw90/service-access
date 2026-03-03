package observer

import (
	"github.com/adityakw90/go-monitoring"
	"github.com/adityakw90/service-access/internal/core/domain/signal"
	"go.opentelemetry.io/otel/attribute"
)

func NewPermissionObserver(
	logger monitoring.Logger,
	tracer monitoring.Tracer,
) *serviceObserver[signal.SignalPermission] {
	return NewServiceObserver(
		logger,
		tracer,
		func(s signal.SignalPermission) []attribute.KeyValue {
			attrs := []attribute.KeyValue{
				attribute.String("operation", s.Operation),
			}
			if s.UID != nil {
				attrs = append(attrs, attribute.String("permission.uid", *s.UID))
			}
			if s.Resource != nil {
				attrs = append(attrs, attribute.String("permission.resource", *s.Resource))
			}
			if s.Action != nil {
				attrs = append(attrs, attribute.String("permission.action", *s.Action))
			}
			if s.Description != nil {
				attrs = append(attrs, attribute.String("permission.description", *s.Description))
			}
			if s.CreatedAt != nil {
				attrs = append(attrs, attribute.Int64("permission.created_at", s.CreatedAt.Unix()))
			}
			if s.UpdatedAt != nil {
				attrs = append(attrs, attribute.Int64("permission.updated_at", s.UpdatedAt.Unix()))
			}
			return attrs
		},
		func(s signal.SignalPermission) map[string]any {
			fields := map[string]any{
				"operation": s.Operation,
			}
			if s.UID != nil {
				fields["permission.uid"] = *s.UID
			}
			if s.Resource != nil {
				fields["permission.resource"] = *s.Resource
			}
			if s.Action != nil {
				fields["permission.action"] = *s.Action
			}
			if s.Description != nil {
				fields["permission.description"] = *s.Description
			}
			if s.CreatedAt != nil {
				fields["permission.created_at"] = *s.CreatedAt
			}
			if s.UpdatedAt != nil {
				fields["permission.updated_at"] = *s.UpdatedAt
			}
			return fields
		},
	)
}
