package resolver

import "context"

type RoleResolver interface {
	IDsByUIDs(ctx context.Context, uids []string) (map[string]int64, error)
	UIDsByIDs(ctx context.Context, ids []int64) (map[int64]string, error)
	Invalidate(ctx context.Context, uids ...string) error
}
