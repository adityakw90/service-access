package repository

import (
	"context"

	"github.com/adityakw90/service-access/internal/core/domain/model"
	"github.com/adityakw90/service-access/internal/core/domain/param"
)

// RoleRepository defines the interface for role repository.
type RoleRepository interface {
	// Manage
	Create(ctx context.Context, role *model.Role) (*model.Role, error) // Create creates a role.
	Update(ctx context.Context, role *model.Role) error                // Update updates a role.
	Delete(ctx context.Context, id int64) error                        // Delete deletes a role.

	// Retrieve
	GetByID(ctx context.Context, id int64) (*model.Role, error)                                                          // GetByID returns a role by ID.
	List(ctx context.Context, pagination *param.PaginationParam, filter *param.RoleListFilterParam) (model.Roles, error) // List returns the list of roles.

	// Permission
	ListPermission(ctx context.Context, roleID int64, pagination *param.PaginationParam, filter *param.RolePermissionListFilterParam) (model.RolePermissions, error)
	AddPermission(ctx context.Context, roleID int64, groupPermissionID int64) error        // AddPermission adds a group permission to a role.
	RemovePermission(ctx context.Context, roleID int64, groupPermissionID int64) error     // RemovePermission removes a group permission from a role.
	ReplacePermission(ctx context.Context, roleID int64, groupPermissionIDs []int64) error // ReplacePermission replaces the group permissions of a role.
}
