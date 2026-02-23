package resolver

import "context"

// InvalidateOpt is a functional option for Invalidate operations.
type InvalidateOpt func(*InvalidateOptions)

// InvalidateOptions holds the options for Invalidate operations.
type InvalidateOptions struct {
	UIDs []string
	IDs  []int64
}

// WithUIDs specifies the UIDs to invalidate.
func WithUIDs(uids ...string) InvalidateOpt {
	return func(o *InvalidateOptions) {
		o.UIDs = uids
	}
}

// WithIDs specifies the IDs to invalidate.
func WithIDs(ids ...int64) InvalidateOpt {
	return func(o *InvalidateOptions) {
		o.IDs = ids
	}
}

type GroupResolver interface {
	IDsByUIDs(ctx context.Context, uids []string) (map[string]int64, error)
	UIDsByIDs(ctx context.Context, ids []int64) (map[int64]string, error)
	Invalidate(ctx context.Context, opts ...InvalidateOpt) error
}
