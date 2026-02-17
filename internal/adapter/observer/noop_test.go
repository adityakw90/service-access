package observer

import (
	"context"
	"testing"

	"github.com/adityakw90/service-access/internal/core/domain/signal"
)

func TestAdapter_Observer_NoopObserver_OnSignal_DoesNotPanic(t *testing.T) {
	noop := NewNoopObserver[signal.SignalAccessCheck]()
	ctx := context.Background()

	// Should not panic
	noop.OnSignal(ctx, signal.SignalStart, signal.SignalAccessCheck{
		SubjectID: "test@example.com",
	}, nil)

	noop.OnSignal(ctx, signal.SignalFail, signal.SignalAccessCheck{
		SubjectID: "test@example.com",
	}, nil)
}

func TestAdapter_Observer_NoopObserver_Generic(t *testing.T) {
	// Test with different types
	authNoop := NewNoopObserver[signal.SignalAccessCheck]()
	userNoop := NewNoopObserver[signal.SignalGroup]()

	ctx := context.Background()

	// Should not panic with any signal type
	authNoop.OnSignal(ctx, signal.SignalSuccess, signal.SignalAccessCheck{}, nil)
	userNoop.OnSignal(ctx, signal.SignalStart, signal.SignalGroup{}, nil)
}
