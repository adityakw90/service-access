package service

import (
	"context"
	"errors"
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
		// Check if it's a not-found error - if so, we have a stale cache entry
		if errors.Is(err, domainerrors.ErrGroupNotFound) {
			// Invalidate the stale resolver mapping
			_ = s.resolvers.Group().Invalidate(ctx, uid)

			s.observer.OnSignal(ctx, signal.SignalReject, signal.SignalGroup{
				UID:       &uid,
				Operation: "get",
			}, err)
			return nil, err
		}

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
	_ = s.resolvers.Group().Invalidate(ctx, uid)

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
	_ = s.resolvers.Group().Invalidate(ctx, uid)

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
	// Resolve group UID
	groupIDs, err := s.resolvers.Group().IDsByUIDs(ctx, []string{groupUID})
	if err != nil {
		s.observer.OnSignal(ctx, signal.SignalError, signal.SignalGroup{
			UID:       &groupUID,
			Operation: "assign_permission",
		}, err)
		return domainerrors.ErrGroupPermissionAssignFailed
	}

	groupID, exists := groupIDs[groupUID]
	if !exists {
		err := domainerrors.ErrGroupNotFound
		s.observer.OnSignal(ctx, signal.SignalReject, signal.SignalGroup{
			UID:       &groupUID,
			Operation: "assign_permission",
		}, err)
		return err
	}

	// Resolve permission UID
	permIDs, err := s.resolvers.Permission().IDsByUIDs(ctx, []string{permissionUID})
	if err != nil {
		s.observer.OnSignal(ctx, signal.SignalError, signal.SignalGroup{
			UID:       &groupUID,
			Operation: "assign_permission",
		}, err)
		return domainerrors.ErrGroupPermissionAssignFailed
	}

	permissionID, exists := permIDs[permissionUID]
	if !exists {
		err := domainerrors.ErrPermissionNotFound
		s.observer.OnSignal(ctx, signal.SignalReject, signal.SignalGroup{
			UID:       &groupUID,
			Operation: "assign_permission",
		}, err)
		return err
	}

	// Generate UID for group permission
	groupPermUID := s.uidGenerator.New()

	// Get group and permission for event
	var group *model.Group
	var permission *model.Permission
	err = s.uow.Do(ctx, func(r repository.RepositoryProvider) error {
		var errUoW error
		group, errUoW = r.Group().GetByID(ctx, groupID)
		if errUoW != nil {
			return errUoW
		}

		permission, errUoW = r.Permission().GetByID(ctx, permissionID)
		if errUoW != nil {
			return errUoW
		}

		// Assign
		if err := r.Group().AddPermission(ctx, groupID, permissionID, groupPermUID); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		s.observer.OnSignal(ctx, signal.SignalError, signal.SignalGroup{
			UID:       &groupUID,
			Operation: "assign_permission",
		}, err)
		return domainerrors.ErrGroupPermissionAssignFailed
	}

	s.publisher.Publish(ctx, event.EventGroupAssignPermission, &event.EventGroupAssignPermissionData{
		GroupUID:      group.UID,
		PermissionUID: permission.UID,
	})

	s.observer.OnSignal(ctx, signal.SignalSuccess, signal.SignalGroup{
		UID:       &groupUID,
		Name:      &group.Name,
		Operation: "assign_permission",
	}, nil)

	return nil
}

func (s *groupService) RevokePermission(ctx context.Context, groupUID string, permissionUID string) error {
	// Resolve group UID
	groupIDs, err := s.resolvers.Group().IDsByUIDs(ctx, []string{groupUID})
	if err != nil {
		s.observer.OnSignal(ctx, signal.SignalError, signal.SignalGroup{
			UID:       &groupUID,
			Operation: "revoke_permission",
		}, err)
		return domainerrors.ErrGroupPermissionRevokeFailed
	}

	groupID, exists := groupIDs[groupUID]
	if !exists {
		err := domainerrors.ErrGroupNotFound
		s.observer.OnSignal(ctx, signal.SignalReject, signal.SignalGroup{
			UID:       &groupUID,
			Operation: "revoke_permission",
		}, err)
		return err
	}

	// Resolve permission UID
	permIDs, err := s.resolvers.Permission().IDsByUIDs(ctx, []string{permissionUID})
	if err != nil {
		s.observer.OnSignal(ctx, signal.SignalError, signal.SignalGroup{
			UID:       &groupUID,
			Operation: "revoke_permission",
		}, err)
		return domainerrors.ErrGroupPermissionRevokeFailed
	}

	permissionID, exists := permIDs[permissionUID]
	if !exists {
		err := domainerrors.ErrPermissionNotFound
		s.observer.OnSignal(ctx, signal.SignalReject, signal.SignalGroup{
			UID:       &groupUID,
			Operation: "revoke_permission",
		}, err)
		return err
	}

	// Get group and permission for event
	var group *model.Group
	var permission *model.Permission
	err = s.uow.Do(ctx, func(r repository.RepositoryProvider) error {
		var errUoW error
		group, errUoW = r.Group().GetByID(ctx, groupID)
		if errUoW != nil {
			return errUoW
		}

		permission, errUoW = r.Permission().GetByID(ctx, permissionID)
		if errUoW != nil {
			return errUoW
		}

		// Revoke
		if err := r.Group().RemovePermission(ctx, groupID, permissionID); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		s.observer.OnSignal(ctx, signal.SignalError, signal.SignalGroup{
			UID:       &groupUID,
			Operation: "revoke_permission",
		}, err)
		return domainerrors.ErrGroupPermissionRevokeFailed
	}

	s.publisher.Publish(ctx, event.EventGroupRevokePermission, &event.EventGroupRevokePermissionData{
		GroupUID:      group.UID,
		PermissionUID: permission.UID,
	})

	s.observer.OnSignal(ctx, signal.SignalSuccess, signal.SignalGroup{
		UID:       &groupUID,
		Name:      &group.Name,
		Operation: "revoke_permission",
	}, nil)

	return nil
}

func (s *groupService) UpdatePermission(ctx context.Context, groupUID string, permissionUIDs []string) error {
	// Resolve group UID
	groupIDs, err := s.resolvers.Group().IDsByUIDs(ctx, []string{groupUID})
	if err != nil {
		s.observer.OnSignal(ctx, signal.SignalError, signal.SignalGroup{
			UID:       &groupUID,
			Operation: "update_permission",
		}, err)
		return domainerrors.ErrGroupPermissionUpdateFailed
	}

	groupID, exists := groupIDs[groupUID]
	if !exists {
		err := domainerrors.ErrGroupNotFound
		s.observer.OnSignal(ctx, signal.SignalReject, signal.SignalGroup{
			UID:       &groupUID,
			Operation: "update_permission",
		}, err)
		return err
	}

	// Batch resolve permission UIDs
	permIDs, err := s.resolvers.Permission().IDsByUIDs(ctx, permissionUIDs)
	if err != nil {
		s.observer.OnSignal(ctx, signal.SignalError, signal.SignalGroup{
			UID:       &groupUID,
			Operation: "update_permission",
		}, err)
		return domainerrors.ErrGroupPermissionUpdateFailed
	}

	// Validate all permissions exist and build ID slice
	permissionIDs := make([]int64, 0, len(permissionUIDs))
	for _, uid := range permissionUIDs {
		id, exists := permIDs[uid]
		if !exists {
			err := domainerrors.ErrPermissionNotFound
			s.observer.OnSignal(ctx, signal.SignalReject, signal.SignalGroup{
				UID:       &groupUID,
				Operation: "update_permission",
			}, err)
			return err
		}
		permissionIDs = append(permissionIDs, id)
	}

	// Generate UIDs
	groupPermUIDs := make([]string, len(permissionUIDs))
	for i := range permissionUIDs {
		groupPermUIDs[i] = s.uidGenerator.New()
	}

	var group *model.Group
	err = s.uow.Do(ctx, func(r repository.RepositoryProvider) error {
		var errUoW error
		group, errUoW = r.Group().GetByID(ctx, groupID)
		if errUoW != nil {
			return errUoW
		}

		// Replace permissions
		if err := r.Group().ReplacePermission(ctx, groupID, permissionIDs, groupPermUIDs); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		s.observer.OnSignal(ctx, signal.SignalError, signal.SignalGroup{
			UID:       &groupUID,
			Operation: "update_permission",
		}, err)
		return domainerrors.ErrGroupPermissionUpdateFailed
	}

	s.publisher.Publish(ctx, event.EventGroupUpdatePermission, &event.EventGroupUpdatePermissionData{
		GroupUID: group.UID,
		UIDs:     permissionUIDs,
	})

	s.observer.OnSignal(ctx, signal.SignalSuccess, signal.SignalGroup{
		UID:       &groupUID,
		Name:      &group.Name,
		Operation: "update_permission",
	}, nil)

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
