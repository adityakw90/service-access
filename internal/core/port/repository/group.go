package repository

import (
	"context"

	"github.com/adityakw90/service-access/internal/core/domain/model"
	"github.com/adityakw90/service-access/internal/core/domain/param"
)

// GroupRepository defines the interface for group repository.
type GroupRepository interface {
	// Manage
	Create(ctx context.Context, group *model.Group) (*model.Group, error) // Create creates a group.
	Update(ctx context.Context, group *model.Group) error                 // Update updates a group.
	Delete(ctx context.Context, id int64) error                           // Delete deletes a group.

	// Retrieve
	GetByID(ctx context.Context, id int64) (*model.Group, error)                                                           // GetByID returns a group by ID.
	List(ctx context.Context, pagination *param.PaginationParam, filter *param.GroupListFilterParam) (model.Groups, error) // List returns the list of groups.

	// Permission
	ListPermission(ctx context.Context, groupID int64, pagination *param.PaginationParam, filter *param.GroupPermissionListFilterParam) (model.GroupPermissions, error)
	AddPermission(ctx context.Context, groupID int64, permissionID int64) error        // AddPermission adds a permission to a group.
	RemovePermission(ctx context.Context, groupID int64, permissionID int64) error     // RemovePermission removes a permission from a group.
	ReplacePermission(ctx context.Context, groupID int64, permissionIDs []int64) error // ReplacePermission replaces the permissions of a group.
}
