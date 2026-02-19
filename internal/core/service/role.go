package service

import (
	"context"
	"fmt"

	"github.com/adityakw90/service-access/internal/core/domain/event"
	"github.com/adityakw90/service-access/internal/core/domain/model"
	"github.com/adityakw90/service-access/internal/core/domain/param"
	portEvent "github.com/adityakw90/service-access/internal/core/port/event"
	"github.com/adityakw90/service-access/internal/core/port/repository"
	"github.com/adityakw90/service-access/internal/core/port/service"
)

type roleService struct {
	uow      repository.UnitOfWork
	repos    repository.RepositoryProvider
	publisher portEvent.EventPublisher
}

func NewRoleService(uow repository.UnitOfWork, repos repository.RepositoryProvider, publisher portEvent.EventPublisher) service.RoleService {
	return &roleService{
		uow:      uow,
		repos:    repos,
		publisher: publisher,
	}
}

func (s *roleService) Create(ctx context.Context, p param.RoleCreateParam) (*model.Role, error) {
	var result *model.Role

	err := s.uow.Do(ctx, func(r repository.Repositories) error {
		role := &model.Role{
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
	role, err := s.repos.Role().GetByUID(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}
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
	var role *model.Role

	err := s.uow.Do(ctx, func(r repository.Repositories) error {
		repo := r.Role()

		var errUoW error
		role, errUoW = repo.GetByUID(ctx, uid)
		if errUoW != nil {
			return fmt.Errorf("failed to get role: %w", errUoW)
		}

		if p.Name != nil {
			role.Name = *p.Name
		}
		if p.Description != nil {
			role.Description = *p.Description
		}

		if err := repo.Update(ctx, role); err != nil {
			return fmt.Errorf("failed to update role: %w", err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	s.publisher.Publish(ctx, event.EventRoleUpdate, &event.EventRoleUpdateData{
		UID:         role.UID,
		Name:        role.Name,
		Description: role.Description,
		UpdatedAt:   role.UpdatedAt,
	})
	return nil
}

func (s *roleService) Delete(ctx context.Context, uid string) error {
	var role *model.Role

	err := s.uow.Do(ctx, func(r repository.Repositories) error {
		repo := r.Role()

		var errUoW error
		role, errUoW = repo.GetByUID(ctx, uid)
		if errUoW != nil {
			return fmt.Errorf("failed to get role: %w", errUoW)
		}

		if err := repo.Delete(ctx, role.ID); err != nil {
			return fmt.Errorf("failed to delete role: %w", err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	s.publisher.Publish(ctx, event.EventRoleDelete, &event.EventRoleDeleteData{
		UID: role.UID,
	})
	return nil
}

func (s *roleService) AssignPermission(ctx context.Context, roleUID string, permissionUID string) error {
	var events []event.Event

	err := s.uow.Do(ctx, func(r repository.Repositories) error {
		role, err := r.Role().GetByUID(ctx, roleUID)
		if err != nil {
			return fmt.Errorf("failed to get role: %w", err)
		}

		// Get the group permission for this role's group and the permission UID
		groupPerm, err := r.Group().GetPermissionByGroupIDAndPermissionUID(ctx, role.GroupID, permissionUID)
		if err != nil {
			return fmt.Errorf("failed to get group permission: %w", err)
		}

		if err := r.Role().AddPermission(ctx, role.ID, groupPerm.ID); err != nil {
			return fmt.Errorf("failed to assign permission: %w", err)
		}

		events = []event.Event{event.NewEventRolePermissionAssigned(role, groupPerm)}
		return nil
	})

	if err != nil {
		return err
	}

	s.publisher.Publish(ctx, events)
	return nil
}

func (s *roleService) RevokePermission(ctx context.Context, roleUID string, permissionUID string) error {
	var events []event.Event

	err := s.uow.Do(ctx, func(r repository.Repositories) error {
		role, err := r.Role().GetByUID(ctx, roleUID)
		if err != nil {
			return fmt.Errorf("failed to get role: %w", err)
		}

		// Get the group permission for this role's group and the permission UID
		groupPerm, err := r.Group().GetPermissionByGroupIDAndPermissionUID(ctx, role.GroupID, permissionUID)
		if err != nil {
			return fmt.Errorf("failed to get group permission: %w", err)
		}

		if err := r.Role().RemovePermission(ctx, role.ID, groupPerm.ID); err != nil {
			return fmt.Errorf("failed to revoke permission: %w", err)
		}

		events = []event.Event{event.NewEventRolePermissionRevoked(role, groupPerm)}
		return nil
	})

	if err != nil {
		return err
	}

	s.publisher.Publish(ctx, events)
	return nil
}

func (s *roleService) UpdatePermission(ctx context.Context, roleUID string, permissionUIDs []string) error {
	var events []event.Event

	err := s.uow.Do(ctx, func(r repository.Repositories) error {
		role, err := r.Role().GetByUID(ctx, roleUID)
		if err != nil {
			return fmt.Errorf("failed to get role: %w", err)
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

		events = []event.Event{event.NewEventRolePermissionsUpdated(role)}
		return nil
	})

	if err != nil {
		return err
	}

	s.publisher.Publish(ctx, events)
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
