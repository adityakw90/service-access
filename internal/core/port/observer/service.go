package observer

import (
	"context"

	"github.com/adityakw90/service-access/internal/core/domain/signal"
)

type ServiceObserver[T any] interface {
	OnSignal(ctx context.Context, signal signal.SignalType, data T, err error)
}
