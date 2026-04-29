package request_test

import (
	"testing"

	"github.com/adityakw90/service-access/internal/adapter/api/grpc/request"
)

func TestAccessRequestFromPb(t *testing.T) {
	req := request.AccessRequestFromPb(nil)
	if req == nil {
		t.Error("AccessRequestFromPb() should not return nil for nil input")
	}
}
