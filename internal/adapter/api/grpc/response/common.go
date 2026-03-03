package response

import (
	"time"

	"github.com/adityakw90/service-access-proto/gen/go/common"
	"github.com/adityakw90/service-access/internal/core/domain/model"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ToProtoMeta converts domain meta to proto meta.
func ToProtoMeta(m *model.Meta) *common.Meta {
	if m == nil {
		return nil
	}
	return &common.Meta{
		Page:  int32(m.Page),
		Limit: int32(m.Limit),
		Total: m.Total,
		Pages: int32(m.Pages),
	}
}

// toProtoTimestamp converts time.Time to protobuf timestamp.
func toProtoTimestamp(t time.Time) *time.Time {
	return &t
}

// toProtoTimestampPB converts time.Time to protobuf timestamp.
func toProtoTimestampPB(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return &timestamppb.Timestamp{
		Seconds: t.Unix(),
		Nanos:   int32(t.Nanosecond()),
	}
}
