package service

import (
	"context"

	"github.com/adityakw90/service-access/internal/core/domain/model"
	"github.com/adityakw90/service-access/internal/core/domain/param"
)

// PermissionService represents the service interface for managing permissions.
type PermissionService interface {
	// retrieve
	List(ctx context.Context, pagination *param.PaginationParam, filter *param.PermissionListFilterParam) (*model.Permissions, error)
	Get(ctx context.Context, uid string) (*model.Permission, error)

	// manage
	Create(ctx context.Context, param param.PermissionCreateParam) (*model.Permission, error)
	Update(ctx context.Context, uid string, param param.PermissionUpdateParam) error
	Delete(ctx context.Context, uid string) error
}

// Constructor signature for PermissionService implementations:
// func NewPermissionService(
//     uow repository.UnitOfWork,
//     repos repository.RepositoryProvider,
//     publisher event.EventPublisher,
//     uidGenerator security.UIDGenerator,
//     resolverProvider resolver.ResolverProvider,
//     obs observer.ServiceObserver[signal.SignalPermission],
// ) PermissionService
