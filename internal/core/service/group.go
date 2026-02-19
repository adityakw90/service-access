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

type groupService struct {
	uow       repository.UnitOfWork
	repos     repository.RepositoryProvider
	publisher portEvent.EventPublisher
}

// NewGroupService creates a new GroupService.
func NewGroupService(uow repository.UnitOfWork, repos repository.RepositoryProvider, publisher portEvent.EventPublisher) service.GroupService {
	return &groupService{
		uow:       uow,
		repos:     repos,
		publisher: publisher,
	}
}

func (s *groupService) Create(ctx context.Context, p param.GroupCreateParam) (*model.Group, error) {
	group := &model.Group{
		Name:        p.Name,
		Description: p.Description,
	}
	err := s.uow.Do(ctx, func(r repository.Repositories) error {
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
	group, err := s.repos.Group().GetByUID(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("failed to get group: %w", err)
	}
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
	var group *model.Group

	err := s.uow.Do(ctx, func(r repository.Repositories) error {
		var errUoW error
		repo := r.Group()

		// Get existing
		group, errUoW = repo.GetByUID(ctx, uid)
		if errUoW != nil {
			return fmt.Errorf("failed to get group: %w", errUoW)
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
			return fmt.Errorf("failed to update group: %w", err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	s.publisher.Publish(ctx, event.EventGroupUpdate, &event.EventGroupUpdateData{
		UID:         group.UID,
		Name:        group.Name,
		Description: group.Description,
		UpdatedAt:   group.UpdatedAt,
	})
	return nil
}

func (s *groupService) Delete(ctx context.Context, uid string) error {
	var events []event.Event

	err := s.uow.Do(ctx, func(r repository.Repositories) error {
		repo := r.Group()

		// Get existing for event
		group, err := repo.GetByUID(ctx, uid)
		if err != nil {
			return fmt.Errorf("failed to get group: %w", err)
		}

		// Delete
		if err := repo.Delete(ctx, group.ID); err != nil {
			return fmt.Errorf("failed to delete group: %w", err)
		}

		events = []event.Event{event.NewEventGroupDeleted(group)}
		return nil
	})

	if err != nil {
		return err
	}

	s.publisher.Publish(ctx, events)
	return nil
}

func (s *groupService) AssignPermission(ctx context.Context, groupUID string, permissionUID string) error {
	var events []event.Event

	err := s.uow.Do(ctx, func(r repository.Repositories) error {
		// Get group
		group, err := r.Group().GetByUID(ctx, groupUID)
		if err != nil {
			return fmt.Errorf("failed to get group: %w", err)
		}

		// Get permission
		permission, err := r.Permission().GetByUID(ctx, permissionUID)
		if err != nil {
			return fmt.Errorf("failed to get permission: %w", err)
		}

		// Assign
		if err := r.Group().AddPermission(ctx, group.ID, permission.ID); err != nil {
			return fmt.Errorf("failed to assign permission: %w", err)
		}

		events = []event.Event{event.NewEventPermissionAssignedToGroup(group, permission)}
		return nil
	})

	if err != nil {
		return err
	}

	s.publisher.Publish(ctx, events)
	return nil
}

func (s *groupService) RevokePermission(ctx context.Context, groupUID string, permissionUID string) error {
	var events []event.Event

	err := s.uow.Do(ctx, func(r repository.Repositories) error {
		// Get group
		group, err := r.Group().GetByUID(ctx, groupUID)
		if err != nil {
			return fmt.Errorf("failed to get group: %w", err)
		}

		// Get permission
		permission, err := r.Permission().GetByUID(ctx, permissionUID)
		if err != nil {
			return fmt.Errorf("failed to get permission: %w", err)
		}

		// Revoke
		if err := r.Group().RemovePermission(ctx, group.ID, permission.ID); err != nil {
			return fmt.Errorf("failed to revoke permission: %w", err)
		}

		events = []event.Event{event.NewEventPermissionRevokedFromGroup(group, permission)}
		return nil
	})

	if err != nil {
		return err
	}

	s.publisher.Publish(ctx, events)
	return nil
}

func (s *groupService) UpdatePermission(ctx context.Context, groupUID string, permissionUIDs []string) error {
	var events []event.Event

	err := s.uow.Do(ctx, func(r repository.Repositories) error {
		// Get group
		group, err := r.Group().GetByUID(ctx, groupUID)
		if err != nil {
			return fmt.Errorf("failed to get group: %w", err)
		}

		// Get all permission IDs
		permissionIDs := make([]int64, 0, len(permissionUIDs))
		for _, uid := range permissionUIDs {
			permission, err := r.Permission().GetByUID(ctx, uid)
			if err != nil {
				return fmt.Errorf("failed to get permission %s: %w", uid, err)
			}
			permissionIDs = append(permissionIDs, permission.ID)
		}

		// Replace permissions
		if err := r.Group().ReplacePermission(ctx, group.ID, permissionIDs); err != nil {
			return fmt.Errorf("failed to replace permissions: %w", err)
		}

		events = []event.Event{event.NewEventGroupPermissionsUpdated(group)}
		return nil
	})

	if err != nil {
		return err
	}

	s.publisher.Publish(ctx, events)
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
