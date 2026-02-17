package resolver

import (
	"context"

	"github.com/adityakw90/service-access/internal/core/domain/param"
)

type GroupPermissionResolver interface {
	IDsByUIDs(ctx context.Context, uids []string) (map[string]int64, error)
	UIDsByIDs(ctx context.Context, ids []int64) (map[int64]string, error)
	IDsByGroupIDAndPermissionIDs(ctx context.Context, param []param.GroupPermissionMapGroupIDPermissionID) (map[param.GroupPermissionMapGroupIDPermissionID]int64, error)
	GroupIDsAndPermissionIDsByIDs(ctx context.Context, ids []int64) (map[int64]param.GroupPermissionMapGroupIDPermissionID, error)
}
