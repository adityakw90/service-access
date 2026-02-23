package service

import (
	"context"

	"github.com/adityakw90/service-access/internal/core/domain/model"
	"github.com/adityakw90/service-access/internal/core/domain/param"
)

// RoleService represents the service interface for managing roles.
type RoleService interface {
	// retrieve
	List(ctx context.Context, pagination *param.PaginationParam, filter *param.RoleListFilterParam) (*model.Roles, error)
	Get(ctx context.Context, uid string) (*model.Role, error)

	// manage
	Create(ctx context.Context, param param.RoleCreateParam) (*model.Role, error)
	Update(ctx context.Context, uid string, param param.RoleUpdateParam) error
	Delete(ctx context.Context, uid string) error

	// permission
	ListPermission(ctx context.Context, roleUID string, pagination *param.PaginationParam, filter *param.RolePermissionListFilterParam) (*model.RolePermissions, error)
	UpdatePermission(ctx context.Context, roleUID string, permissionUID []string) error
	AssignPermission(ctx context.Context, roleUID string, permissionUID string) error
	RevokePermission(ctx context.Context, roleUID string, permissionUID string) error
}

// Constructor signature for RoleService implementations:
// func NewRoleService(
//     uow repository.UnitOfWork,
//     repos repository.RepositoryProvider,
//     publisher event.EventPublisher,
//     uidGenerator security.UIDGenerator,
//     resolverProvider resolver.ResolverProvider,
//     obs observer.ServiceObserver[signal.SignalRole],
// ) RoleService

