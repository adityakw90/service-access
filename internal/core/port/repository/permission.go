package repository

import (
	"context"

	"github.com/adityakw90/service-access/internal/core/domain/model"
	"github.com/adityakw90/service-access/internal/core/domain/param"
)

// PermissionRepository defines the interface for permission repository.
type PermissionRepository interface {
	// Manage
	Create(ctx context.Context, permission *model.Permission) error
	Update(ctx context.Context, permission *model.Permission) error
	Delete(ctx context.Context, id int64) error

	// Retrieve
	GetByID(ctx context.Context, id int64) (*model.Permission, error)
	List(ctx context.Context, pagination *param.PaginationParam, filter *param.PermissionListFilterParam) (model.Permissions, error)
}
