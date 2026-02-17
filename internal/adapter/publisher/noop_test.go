package publisher

import (
	"context"
	"testing"

	"github.com/adityakw90/service-access/internal/core/domain/event"
)

func TestNoOpPublisher(t *testing.T) {
	pub := NewNoOpPublisher()

	ctx := context.Background()

	// Should not return error
	err := pub.Publish(ctx, event.EventAccessCheck, event.EventAccessCheckData{
		SubjectId:   "uid-from-user-service",
		SubjectType: "user",
		Resource:    "user",
		Action:      "read",
		Reason:      "user not have any permission",
	})
	if err != nil {
		t.Errorf("NoOpPublisher.Publish() error = %v", err)
	}

	// Close should not return error
	err = pub.Close()
	if err != nil {
		t.Errorf("NoOpPublisher.Close() error = %v", err)
	}
}
