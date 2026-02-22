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

type groupService struct {
	uow         repository.UnitOfWork
	repos       repository.RepositoryProvider
	publisher   portEvent.EventPublisher
	uidGenerator security.UIDGenerator
	resolvers   resolver.ResolverProvider
	observer    observer.ServiceObserver[signal.SignalGroup]
}

// NewGroupService creates a new GroupService.
func NewGroupService(
	uow repository.UnitOfWork,
	repos repository.RepositoryProvider,
	publisher portEvent.EventPublisher,
	uidGenerator security.UIDGenerator,
	resolverProvider resolver.ResolverProvider,
	observer observer.ServiceObserver[signal.SignalGroup],
) service.GroupService {
	return &groupService{
		uow:         uow,
		repos:       repos,
		publisher:   publisher,
		uidGenerator: uidGenerator,
		resolvers:   resolverProvider,
		observer:    observer,
	}
}

func (s *groupService) Create(ctx context.Context, p param.GroupCreateParam) (*model.Group, error) {
	group := &model.Group{
		UID:         s.uidGenerator.New(),
		Name:        p.Name,
		Description: p.Description,
	}
	err := s.uow.Do(ctx, func(r repository.RepositoryProvider) error {
		if err := r.Group().Create(ctx, group); err != nil {
			return fmt.Errorf("failed to create group: %w", err)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	s.publisher.Publish(ctx, event.EventGroupCreate, &event.EventGroupCreateData{
		UID:         group.UID,
		Name:        group.Name,
		Description: group.Description,
		CreatedAt:   group.CreatedAt,
	})
	return group, nil
}

func (s *groupService) Get(ctx context.Context, uid string) (*model.Group, error) {
	ids, err := s.resolvers.Group().IDsByUIDs(ctx, []string{uid})
	if err != nil {
		s.observer.OnSignal(ctx, signal.SignalError, signal.SignalGroup{
			UID:       &uid,
			Operation: "get",
		}, err)
		return nil, domainerrors.ErrGroupGetFailed
	}

	id, exists := ids[uid]
	if !exists {
		err := domainerrors.ErrGroupNotFound
		s.observer.OnSignal(ctx, signal.SignalReject, signal.SignalGroup{
			UID:       &uid,
			Operation: "get",
		}, err)
		return nil, err
	}

	group, err := s.repos.Group().GetByID(ctx, id)
	if err != nil {
		s.observer.OnSignal(ctx, signal.SignalError, signal.SignalGroup{
			UID:       &uid,
			Operation: "get",
		}, err)
		return nil, domainerrors.ErrGroupGetFailed
	}

	s.observer.OnSignal(ctx, signal.SignalSuccess, signal.SignalGroup{
		UID:       &uid,
		Name:      &group.Name,
		Operation: "get",
	}, nil)

	return group, nil
}

func (s *groupService) List(ctx context.Context, pagination *param.PaginationParam, filter *param.GroupListFilterParam) (*model.Groups, error) {
	groups, err := s.repos.Group().List(ctx, pagination, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list groups: %w", err)
	}
	return &groups, nil
}

func (s *groupService) Update(ctx context.Context, uid string, p param.GroupUpdateParam) error {
	// Resolve UID before transaction
	ids, err := s.resolvers.Group().IDsByUIDs(ctx, []string{uid})
	if err != nil {
		s.observer.OnSignal(ctx, signal.SignalError, signal.SignalGroup{
			UID:       &uid,
			Operation: "update",
		}, err)
		return domainerrors.ErrGroupUpdateFailed
	}

	id, exists := ids[uid]
	if !exists {
		err := domainerrors.ErrGroupNotFound
		s.observer.OnSignal(ctx, signal.SignalReject, signal.SignalGroup{
			UID:       &uid,
			Operation: "update",
		}, err)
		return err
	}

	var group *model.Group
	err = s.uow.Do(ctx, func(r repository.RepositoryProvider) error {
		var errUoW error
		repo := r.Group()

		// Get existing by ID
		group, errUoW = repo.GetByID(ctx, id)
		if errUoW != nil {
			return errUoW
		}

		// Update fields
		if p.Name != nil {
			group.Name = *p.Name
		}
		if p.Description != nil {
			group.Description = *p.Description
		}

		// Persist
		if err := repo.Update(ctx, group); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		s.observer.OnSignal(ctx, signal.SignalError, signal.SignalGroup{
			UID:       &uid,
			Operation: "update",
		}, err)
		return domainerrors.ErrGroupUpdateFailed
	}

	// Invalidate resolver cache
	// TODO: Add Invalidate call after Task 21 (port interface update)
	// _ = s.resolvers.Group().Invalidate(ctx, uid)

	s.publisher.Publish(ctx, event.EventGroupUpdate, &event.EventGroupUpdateData{
		UID:         group.UID,
		Name:        group.Name,
		Description: group.Description,
		UpdatedAt:   group.UpdatedAt,
	})

	s.observer.OnSignal(ctx, signal.SignalSuccess, signal.SignalGroup{
		UID:       &uid,
		Name:      &group.Name,
		Operation: "update",
	}, nil)

	return nil
}

func (s *groupService) Delete(ctx context.Context, uid string) error {
	// Resolve UID before transaction
	ids, err := s.resolvers.Group().IDsByUIDs(ctx, []string{uid})
	if err != nil {
		s.observer.OnSignal(ctx, signal.SignalError, signal.SignalGroup{
			UID:       &uid,
			Operation: "delete",
		}, err)
		return domainerrors.ErrGroupDeleteFailed
	}

	id, exists := ids[uid]
	if !exists {
		err := domainerrors.ErrGroupNotFound
		s.observer.OnSignal(ctx, signal.SignalReject, signal.SignalGroup{
			UID:       &uid,
			Operation: "delete",
		}, err)
		return err
	}

	var group *model.Group
	err = s.uow.Do(ctx, func(r repository.RepositoryProvider) error {
		repo := r.Group()

		// Get existing for event
		var errUoW error
		group, errUoW = repo.GetByID(ctx, id)
		if errUoW != nil {
			return errUoW
		}

		// Delete
		if err := repo.Delete(ctx, id); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		s.observer.OnSignal(ctx, signal.SignalError, signal.SignalGroup{
			UID:       &uid,
			Operation: "delete",
		}, err)
		return domainerrors.ErrGroupDeleteFailed
	}

	// Invalidate resolver cache
	// TODO: Add Invalidate call after Task 21 (port interface update)
	// _ = s.resolvers.Group().Invalidate(ctx, uid)

	s.publisher.Publish(ctx, event.EventGroupDelete, &event.EventGroupDeleteData{
		UID: group.UID,
	})

	s.observer.OnSignal(ctx, signal.SignalSuccess, signal.SignalGroup{
		UID:       &uid,
		Name:      &group.Name,
		Operation: "delete",
	}, nil)

	return nil
}

func (s *groupService) AssignPermission(ctx context.Context, groupUID string, permissionUID string) error {
	var group *model.Group
	var permission *model.Permission

	err := s.uow.Do(ctx, func(r repository.RepositoryProvider) error {
		var errUoW error
		// Get group
		group, errUoW = r.Group().GetByUID(ctx, groupUID)
		if errUoW != nil {
			return fmt.Errorf("failed to get group: %w", errUoW)
		}

		// Get permission
		permission, errUoW = r.Permission().GetByUID(ctx, permissionUID)
		if errUoW != nil {
			return fmt.Errorf("failed to get permission: %w", errUoW)
		}

		// Generate UID for group permission
		groupPermUID := s.uidGenerator.New()

		// Assign
		if err := r.Group().AddPermission(ctx, group.ID, permission.ID, groupPermUID); err != nil {
			return fmt.Errorf("failed to assign permission: %w", err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	s.publisher.Publish(ctx, event.EventGroupAssignPermission, &event.EventGroupAssignPermissionData{
		GroupUID:      group.UID,
		PermissionUID: permission.UID,
	})
	return nil
}

func (s *groupService) RevokePermission(ctx context.Context, groupUID string, permissionUID string) error {
	var group *model.Group
	var permission *model.Permission

	err := s.uow.Do(ctx, func(r repository.RepositoryProvider) error {
		var errUoW error
		// Get group
		group, errUoW = r.Group().GetByUID(ctx, groupUID)
		if errUoW != nil {
			return fmt.Errorf("failed to get group: %w", errUoW)
		}

		// Get permission
		permission, errUoW = r.Permission().GetByUID(ctx, permissionUID)
		if errUoW != nil {
			return fmt.Errorf("failed to get permission: %w", errUoW)
		}

		// Revoke
		if err := r.Group().RemovePermission(ctx, group.ID, permission.ID); err != nil {
			return fmt.Errorf("failed to revoke permission: %w", err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	s.publisher.Publish(ctx, event.EventGroupRevokePermission, &event.EventGroupRevokePermissionData{
		GroupUID:      group.UID,
		PermissionUID: permission.UID,
	})
	return nil
}

func (s *groupService) UpdatePermission(ctx context.Context, groupUID string, permissionUIDs []string) error {
	var group *model.Group

	err := s.uow.Do(ctx, func(r repository.RepositoryProvider) error {
		var errUoW error
		// Get group
		group, errUoW = r.Group().GetByUID(ctx, groupUID)
		if errUoW != nil {
			return fmt.Errorf("failed to get group: %w", errUoW)
		}

		// Get all permission IDs and generate UIDs for each
		permissionIDs := make([]int64, 0, len(permissionUIDs))
		groupPermUIDs := make([]string, 0, len(permissionUIDs))
		for _, uid := range permissionUIDs {
			permission, err := r.Permission().GetByUID(ctx, uid)
			if err != nil {
				return fmt.Errorf("failed to get permission %s: %w", uid, err)
			}
			permissionIDs = append(permissionIDs, permission.ID)
			groupPermUIDs = append(groupPermUIDs, s.uidGenerator.New())
		}

		// Replace permissions
		if err := r.Group().ReplacePermission(ctx, group.ID, permissionIDs, groupPermUIDs); err != nil {
			return fmt.Errorf("failed to replace permissions: %w", err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	s.publisher.Publish(ctx, event.EventGroupUpdatePermission, &event.EventGroupUpdatePermissionData{
		GroupUID: group.UID,
		UIDs:     permissionUIDs,
	})
	return nil
}

func (s *groupService) ListPermission(ctx context.Context, groupUID string, pagination *param.PaginationParam, filter *param.GroupPermissionListFilterParam) (*model.GroupPermissions, error) {
	// First get the group to validate it exists and get the ID
	group, err := s.repos.Group().GetByUID(ctx, groupUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get group: %w", err)
	}

	// List permissions
	permissions, err := s.repos.Group().ListPermission(ctx, group.ID, pagination, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions: %w", err)
	}

	return &permissions, nil
}
