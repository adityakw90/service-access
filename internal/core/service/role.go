package service

import (
	"context"
	"fmt"

	domainerrors "github.com/adityakw90/service-access/internal/core/domain/errors"
	"github.com/adityakw90/service-access/internal/core/domain/event"
	"github.com/adityakw90/service-access/internal/core/domain/model"
	"github.com/adityakw90/service-access/internal/core/domain/param"
	"github.com/adityakw90/service-access/internal/core/domain/signal"
	portEvent "github.com/adityakw90/service-access/internal/core/port/event"
	"github.com/adityakw90/service-access/internal/core/port/observer"
	"github.com/adityakw90/service-access/internal/core/port/repository"
	"github.com/adityakw90/service-access/internal/core/port/resolver"
	"github.com/adityakw90/service-access/internal/core/port/security"
	"github.com/adityakw90/service-access/internal/core/port/service"
)

type roleService struct {
	uow         repository.UnitOfWork
	repos       repository.RepositoryProvider
	publisher   portEvent.EventPublisher
	uidGenerator security.UIDGenerator
	resolvers   resolver.ResolverProvider
	observer    observer.ServiceObserver[signal.SignalRole]
}

// NewRoleService creates a new RoleService.
func NewRoleService(
	uow repository.UnitOfWork,
	repos repository.RepositoryProvider,
	publisher portEvent.EventPublisher,
	uidGenerator security.UIDGenerator,
	resolverProvider resolver.ResolverProvider,
	observer observer.ServiceObserver[signal.SignalRole],
) service.RoleService {
	return &roleService{
		uow:         uow,
		repos:       repos,
		publisher:   publisher,
		uidGenerator: uidGenerator,
		resolvers:   resolverProvider,
		observer:    observer,
	}
}

func (s *roleService) Create(ctx context.Context, p param.RoleCreateParam) (*model.Role, error) {
	var result *model.Role

	err := s.uow.Do(ctx, func(r repository.RepositoryProvider) error {
		role := &model.Role{
			UID:         s.uidGenerator.New(),
			GroupID:     p.GroupID,
			Name:        p.Name,
			Description: p.Description,
		}
		if err := r.Role().Create(ctx, role); err != nil {
			return fmt.Errorf("failed to create role: %w", err)
		}
		result = role
		return nil
	})

	if err != nil {
		return nil, err
	}

	s.publisher.Publish(ctx, event.EventRoleCreate, &event.EventRoleCreateData{
		UID:         result.UID,
		GroupUID:    result.GroupUID,
		Name:        result.Name,
		Description: result.Description,
		CreatedAt:   result.CreatedAt,
	})
	return result, nil
}

func (s *roleService) Get(ctx context.Context, uid string) (*model.Role, error) {
	ids, err := s.resolvers.Role().IDsByUIDs(ctx, []string{uid})
	if err != nil {
		s.observer.OnSignal(ctx, signal.SignalError, signal.SignalRole{
			UID:       &uid,
			Operation: "get",
		}, err)
		return nil, domainerrors.ErrRoleGetFailed
	}

	id, exists := ids[uid]
	if !exists {
		err := domainerrors.ErrRoleNotFound
		s.observer.OnSignal(ctx, signal.SignalReject, signal.SignalRole{
			UID:       &uid,
			Operation: "get",
		}, err)
		return nil, err
	}

	role, err := s.repos.Role().GetByID(ctx, id)
	if err != nil {
		s.observer.OnSignal(ctx, signal.SignalError, signal.SignalRole{
			UID:       &uid,
			Operation: "get",
		}, err)
		return nil, domainerrors.ErrRoleGetFailed
	}

	s.observer.OnSignal(ctx, signal.SignalSuccess, signal.SignalRole{
		UID:       &uid,
		Name:      &role.Name,
		Operation: "get",
	}, nil)

	return role, nil
}

func (s *roleService) List(ctx context.Context, pagination *param.PaginationParam, filter *param.RoleListFilterParam) (*model.Roles, error) {
	roles, err := s.repos.Role().List(ctx, pagination, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}
	return &roles, nil
}

func (s *roleService) Update(ctx context.Context, uid string, p param.RoleUpdateParam) error {
	// Resolve UID before transaction
	ids, err := s.resolvers.Role().IDsByUIDs(ctx, []string{uid})
	if err != nil {
		s.observer.OnSignal(ctx, signal.SignalError, signal.SignalRole{
			UID:       &uid,
			Operation: "update",
		}, err)
		return domainerrors.ErrRoleUpdateFailed
	}

	id, exists := ids[uid]
	if !exists {
		err := domainerrors.ErrRoleNotFound
		s.observer.OnSignal(ctx, signal.SignalReject, signal.SignalRole{
			UID:       &uid,
			Operation: "update",
		}, err)
		return err
	}

	var role *model.Role
	err = s.uow.Do(ctx, func(r repository.RepositoryProvider) error {
		repo := r.Role()

		var errUoW error
		role, errUoW = repo.GetByID(ctx, id)
		if errUoW != nil {
			return errUoW
		}

		if p.Name != nil {
			role.Name = *p.Name
		}
		if p.Description != nil {
			role.Description = *p.Description
		}

		if err := repo.Update(ctx, role); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		s.observer.OnSignal(ctx, signal.SignalError, signal.SignalRole{
			UID:       &uid,
			Operation: "update",
		}, err)
		return domainerrors.ErrRoleUpdateFailed
	}

	// Invalidate resolver cache
	// TODO: Add Invalidate call after Task 21 (port interface update)
	// _ = s.resolvers.Role().Invalidate(ctx, uid)

	s.publisher.Publish(ctx, event.EventRoleUpdate, &event.EventRoleUpdateData{
		UID:         role.UID,
		Name:        role.Name,
		Description: role.Description,
		UpdatedAt:   role.UpdatedAt,
	})

	s.observer.OnSignal(ctx, signal.SignalSuccess, signal.SignalRole{
		UID:       &uid,
		Name:      &role.Name,
		Operation: "update",
	}, nil)

	return nil
}

func (s *roleService) Delete(ctx context.Context, uid string) error {
	// Resolve UID before transaction
	ids, err := s.resolvers.Role().IDsByUIDs(ctx, []string{uid})
	if err != nil {
		s.observer.OnSignal(ctx, signal.SignalError, signal.SignalRole{
			UID:       &uid,
			Operation: "delete",
		}, err)
		return domainerrors.ErrRoleDeleteFailed
	}

	id, exists := ids[uid]
	if !exists {
		err := domainerrors.ErrRoleNotFound
		s.observer.OnSignal(ctx, signal.SignalReject, signal.SignalRole{
			UID:       &uid,
			Operation: "delete",
		}, err)
		return err
	}

	var role *model.Role
	err = s.uow.Do(ctx, func(r repository.RepositoryProvider) error {
		repo := r.Role()

		var errUoW error
		role, errUoW = repo.GetByID(ctx, id)
		if errUoW != nil {
			return errUoW
		}

		if err := repo.Delete(ctx, id); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		s.observer.OnSignal(ctx, signal.SignalError, signal.SignalRole{
			UID:       &uid,
			Operation: "delete",
		}, err)
		return domainerrors.ErrRoleDeleteFailed
	}

	// Invalidate resolver cache
	// TODO: Add Invalidate call after Task 21 (port interface update)
	// _ = s.resolvers.Role().Invalidate(ctx, uid)

	s.publisher.Publish(ctx, event.EventRoleDelete, &event.EventRoleDeleteData{
		UID: role.UID,
	})

	s.observer.OnSignal(ctx, signal.SignalSuccess, signal.SignalRole{
		UID:       &uid,
		Name:      &role.Name,
		Operation: "delete",
	}, nil)

	return nil
}

func (s *roleService) AssignPermission(ctx context.Context, roleUID string, permissionUID string) error {
	var role *model.Role
	var groupPerm *model.GroupPermission

	err := s.uow.Do(ctx, func(r repository.RepositoryProvider) error {
		var errUoW error
		role, errUoW = r.Role().GetByUID(ctx, roleUID)
		if errUoW != nil {
			return fmt.Errorf("failed to get role: %w", errUoW)
		}

		// Get the group permission for this role's group and the permission UID
		groupPerm, errUoW = r.Group().GetPermissionByGroupIDAndPermissionUID(ctx, role.GroupID, permissionUID)
		if errUoW != nil {
			return fmt.Errorf("failed to get group permission: %w", errUoW)
		}

		if err := r.Role().AddPermission(ctx, role.ID, groupPerm.ID); err != nil {
			return fmt.Errorf("failed to assign permission: %w", err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	s.publisher.Publish(ctx, event.EventRoleAssignPermission, &event.EventRoleAssignPermissionData{
		RoleUID:            role.UID,
		GroupPermissionUID: groupPerm.UID,
	})
	return nil
}

func (s *roleService) RevokePermission(ctx context.Context, roleUID string, permissionUID string) error {
	var role *model.Role
	var groupPerm *model.GroupPermission

	err := s.uow.Do(ctx, func(r repository.RepositoryProvider) error {
		var errUoW error
		role, errUoW = r.Role().GetByUID(ctx, roleUID)
		if errUoW != nil {
			return fmt.Errorf("failed to get role: %w", errUoW)
		}

		// Get the group permission for this role's group and the permission UID
		groupPerm, errUoW = r.Group().GetPermissionByGroupIDAndPermissionUID(ctx, role.GroupID, permissionUID)
		if errUoW != nil {
			return fmt.Errorf("failed to get group permission: %w", errUoW)
		}

		if err := r.Role().RemovePermission(ctx, role.ID, groupPerm.ID); err != nil {
			return fmt.Errorf("failed to revoke permission: %w", err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	s.publisher.Publish(ctx, event.EventRoleRevokePermission, &event.EventRoleRevokePermissionData{
		RoleUID:            role.UID,
		GroupPermissionUID: groupPerm.UID,
	})
	return nil
}

func (s *roleService) UpdatePermission(ctx context.Context, roleUID string, permissionUIDs []string) error {
	var role *model.Role

	err := s.uow.Do(ctx, func(r repository.RepositoryProvider) error {
		var errUoW error
		role, errUoW = r.Role().GetByUID(ctx, roleUID)
		if errUoW != nil {
			return fmt.Errorf("failed to get role: %w", errUoW)
		}

		groupPermissionIDs := make([]int64, 0, len(permissionUIDs))
		for _, uid := range permissionUIDs {
			groupPerm, err := r.Group().GetPermissionByGroupIDAndPermissionUID(ctx, role.GroupID, uid)
			if err != nil {
				return fmt.Errorf("failed to get group permission for permission %s: %w", uid, err)
			}
			groupPermissionIDs = append(groupPermissionIDs, groupPerm.ID)
		}

		if err := r.Role().ReplacePermission(ctx, role.ID, groupPermissionIDs); err != nil {
			return fmt.Errorf("failed to replace permissions: %w", err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	s.publisher.Publish(ctx, event.EventRoleUpdatePermission, &event.EventRoleUpdatePermissionData{
		RoleUID: role.UID,
		UIDs:    permissionUIDs,
	})
	return nil
}

func (s *roleService) ListPermission(ctx context.Context, roleUID string, pagination *param.PaginationParam, filter *param.RolePermissionListFilterParam) (*model.RolePermissions, error) {
	// First get the role to validate it exists and get the ID
	role, err := s.repos.Role().GetByUID(ctx, roleUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	// List permissions
	permissions, err := s.repos.Role().ListPermission(ctx, role.ID, pagination, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions: %w", err)
	}

	return &permissions, nil
}
