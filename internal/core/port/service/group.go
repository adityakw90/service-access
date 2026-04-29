package service

import (
	"context"

	"github.com/adityakw90/service-access/internal/core/domain/model"
	"github.com/adityakw90/service-access/internal/core/domain/param"
)

// GroupService represents the service interface for managing groups.
type GroupService interface {
	// retrieve
	List(ctx context.Context, pagination *param.PaginationParam, filter *param.GroupListFilterParam) (*model.Groups, error) // List retrieves a list of groups with pagination.
	Get(ctx context.Context, uid string) (*model.Group, error)                                                              // Get retrieves a group by its UID.

	// manage
	Create(ctx context.Context, param param.GroupCreateParam) (*model.Group, error) // Create creates a new group.
	Update(ctx context.Context, uid string, param param.GroupUpdateParam) error     // Update updates an existing group.
	Delete(ctx context.Context, uid string) error                                   // Delete deletes a group by its UID.

	// permission
	ListPermission(ctx context.Context, groupUID string, pagination *param.PaginationParam, filter *param.GroupPermissionListFilterParam) (*model.GroupPermissions, error) // ListPermission lists permissions for a group.
	UpdatePermission(ctx context.Context, groupUID string, permissionUID []string) error                                                                                   // UpdatePermission replaces all permissions for a group.
	AssignPermission(ctx context.Context, groupUID string, permissionUID string) error                                                                                     // AssignPermission adds a permission to a group.
	RevokePermission(ctx context.Context, groupUID string, permissionUID string) error                                                                                     // RevokePermission removes a permission from a group.
}

// Constructor signature for GroupService implementations:
// func NewGroupService(
//     uow repository.UnitOfWork,
//     repos repository.RepositoryProvider,
//     publisher event.EventPublisher,
//     uidGenerator security.UIDGenerator,
//     resolverProvider resolver.ResolverProvider,
//     obs observer.ServiceObserver[signal.SignalGroup],
// ) GroupService
