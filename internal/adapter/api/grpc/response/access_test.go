package response_test

import (
	"testing"

	"github.com/adityakw90/service-access/internal/adapter/api/grpc/response"
)

func TestToAccessResponse(t *testing.T) {
	resp := response.ToAccessResponse(true, "test reason")
	if resp == nil {
		t.Error("ToAccessResponse() should not return nil")
	}
	if !resp.Allowed {
		t.Error("ToAccessResponse() allowed should be true")
	}
	if resp.Reason != "test reason" {
		t.Errorf("ToAccessResponse() reason = %v, want 'test reason'", resp.Reason)
	}
}
