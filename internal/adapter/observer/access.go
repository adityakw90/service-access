package observer

import (
	"github.com/adityakw90/go-monitoring"
	"github.com/adityakw90/service-access/internal/core/domain/signal"
	"go.opentelemetry.io/otel/attribute"
)

func NewAccessObserver(
	logger monitoring.Logger,
	tracer monitoring.Tracer,
) *serviceObserver[signal.SignalAccessCheck] {
	return NewServiceObserver(
		logger,
		tracer,
		func(s signal.SignalAccessCheck) []attribute.KeyValue {
			attrs := []attribute.KeyValue{
				attribute.String("subject_id", s.SubjectID),
				attribute.String("subject_type", s.SubjectType),
				attribute.String("resource", s.Resource),
				attribute.String("action", s.Action),
			}
			if s.Allowed != nil {
				attrs = append(attrs, attribute.Bool("allowed", *s.Allowed))
			}
			if s.Reason != nil {
				attrs = append(attrs, attribute.String("reason", *s.Reason))
			}
			return attrs
		},
		func(s signal.SignalAccessCheck) map[string]any {
			fields := map[string]any{
				"subject_id":   s.SubjectID,
				"subject_type": s.SubjectType,
				"resource":     s.Resource,
				"action":       s.Action,
			}
			if s.Allowed != nil {
				fields["allowed"] = *s.Allowed
			}
			if s.Reason != nil {
				fields["reason"] = *s.Reason
			}
			return fields
		},
	)
}
