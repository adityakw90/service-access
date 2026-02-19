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

type permissionService struct {
	uow       repository.UnitOfWork
	repos     repository.RepositoryProvider
	publisher portEvent.EventPublisher
}

func NewPermissionService(uow repository.UnitOfWork, repos repository.RepositoryProvider, publisher portEvent.EventPublisher) service.PermissionService {
	return &permissionService{
		uow:       uow,
		repos:     repos,
		publisher: publisher,
	}
}

func (s *permissionService) Create(ctx context.Context, p param.PermissionCreateParam) (*model.Permission, error) {
	var result *model.Permission
	var events []event.Event

	err := s.uow.Do(ctx, func(r repository.Repositories) error {
		permission := &model.Permission{
			Resource:    p.Resource,
			Action:      p.Action,
			Description: p.Description,
		}
		if err := r.Permission().Create(ctx, permission); err != nil {
			return fmt.Errorf("failed to create permission: %w", err)
		}
		result = permission
		events = []event.Event{event.NewEventPermissionCreated(permission)}
		return nil
	})

	if err != nil {
		return nil, err
	}

	s.publisher.Publish(ctx, events)
	return result, nil
}

func (s *permissionService) Get(ctx context.Context, uid string) (*model.Permission, error) {
	permission, err := s.repos.Permission().GetByUID(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}
	return permission, nil
}

func (s *permissionService) List(ctx context.Context, pagination *param.PaginationParam, filter *param.PermissionListFilterParam) (*model.Permissions, error) {
	permissions, err := s.repos.Permission().List(ctx, pagination, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions: %w", err)
	}
	return &permissions, nil
}

func (s *permissionService) Update(ctx context.Context, uid string, p param.PermissionUpdateParam) error {
	var events []event.Event

	err := s.uow.Do(ctx, func(r repository.Repositories) error {
		repo := r.Permission()

		permission, err := repo.GetByUID(ctx, uid)
		if err != nil {
			return fmt.Errorf("failed to get permission: %w", err)
		}

		if p.Resource != nil {
			permission.Resource = *p.Resource
		}
		if p.Action != nil {
			permission.Action = *p.Action
		}
		if p.Description != nil {
			permission.Description = *p.Description
		}

		if err := repo.Update(ctx, permission); err != nil {
			return fmt.Errorf("failed to update permission: %w", err)
		}

		events = []event.Event{event.NewEventPermissionUpdated(permission)}
		return nil
	})

	if err != nil {
		return err
	}

	s.publisher.Publish(ctx, events)
	return nil
}

func (s *permissionService) Delete(ctx context.Context, uid string) error {
	var events []event.Event

	err := s.uow.Do(ctx, func(r repository.Repositories) error {
		repo := r.Permission()

		permission, err := repo.GetByUID(ctx, uid)
		if err != nil {
			return fmt.Errorf("failed to get permission: %w", err)
		}

		if err := repo.Delete(ctx, permission.ID); err != nil {
			return fmt.Errorf("failed to delete permission: %w", err)
		}

		events = []event.Event{event.NewEventPermissionDeleted(permission)}
		return nil
	})

	if err != nil {
		return err
	}

	s.publisher.Publish(ctx, events)
	return nil
}
