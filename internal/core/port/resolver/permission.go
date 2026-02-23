package resolver

import (
	"context"

	"github.com/adityakw90/service-access/internal/core/domain/param"
)

type PermissionResolver interface {
	IDsByUIDs(ctx context.Context, uids []string) (map[string]int64, error)
	UIDsByIDs(ctx context.Context, ids []int64) (map[int64]string, error)
	IDsByResourceActions(ctx context.Context, resourceActions []param.PermissionMapResourceAction) (map[param.PermissionMapResourceAction]int64, error)
	ResourceActionsByIDs(ctx context.Context, ids []int64) (map[int64]param.PermissionMapResourceAction, error)
	Invalidate(ctx context.Context, uids ...string) error
	InvalidateByIDs(ctx context.Context, ids ...int64) error
}
