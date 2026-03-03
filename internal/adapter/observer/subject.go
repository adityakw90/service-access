package observer

import (
	"github.com/adityakw90/go-monitoring"
	"github.com/adityakw90/service-access/internal/core/domain/signal"
	"go.opentelemetry.io/otel/attribute"
)

func NewSubjectObserver(
	logger monitoring.Logger,
	tracer monitoring.Tracer,
) *serviceObserver[signal.SignalSubject] {
	return NewServiceObserver(
		logger,
		tracer,
		func(s signal.SignalSubject) []attribute.KeyValue {
			attrs := []attribute.KeyValue{
				attribute.String("operation", s.Operation),
			}
			if s.SubjectID != nil {
				attrs = append(attrs, attribute.String("subject_id", *s.SubjectID))
			}
			if s.SubjectType != nil {
				attrs = append(attrs, attribute.String("subject_type", *s.SubjectType))
			}
			if s.RoleUID != nil {
				attrs = append(attrs, attribute.String("role_uid", *s.RoleUID))
			}
			if s.AssignedAt != nil {
				attrs = append(attrs, attribute.Int64("assigned_at", s.AssignedAt.Unix()))
			}
			return attrs
		},
		func(s signal.SignalSubject) map[string]any {
			fields := map[string]any{
				"operation": s.Operation,
			}
			if s.SubjectID != nil {
				fields["subject_id"] = *s.SubjectID
			}
			if s.SubjectType != nil {
				fields["subject_type"] = *s.SubjectType
			}
			if s.RoleUID != nil {
				fields["role_uid"] = *s.RoleUID
			}
			if s.AssignedAt != nil {
				fields["assigned_at"] = *s.AssignedAt
			}
			return fields
		},
	)
}
