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

type permissionService struct {
	uow         repository.UnitOfWork
	repos       repository.RepositoryProvider
	publisher   portEvent.EventPublisher
	uidGenerator security.UIDGenerator
	resolvers   resolver.ResolverProvider
	observer    observer.ServiceObserver[signal.SignalPermission]
}

func NewPermissionService(
	uow repository.UnitOfWork,
	repos repository.RepositoryProvider,
	publisher portEvent.EventPublisher,
	uidGenerator security.UIDGenerator,
	resolverProvider resolver.ResolverProvider,
	observer observer.ServiceObserver[signal.SignalPermission],
) service.PermissionService {
	return &permissionService{
		uow:         uow,
		repos:       repos,
		publisher:   publisher,
		uidGenerator: uidGenerator,
		resolvers:   resolverProvider,
		observer:    observer,
	}
}

func (s *permissionService) Create(ctx context.Context, p param.PermissionCreateParam) (*model.Permission, error) {
	var result *model.Permission

	err := s.uow.Do(ctx, func(r repository.RepositoryProvider) error {
		permission := &model.Permission{
			UID:         s.uidGenerator.New(),
			Resource:    p.Resource,
			Action:      p.Action,
			Description: p.Description,
		}
		if err := r.Permission().Create(ctx, permission); err != nil {
			return fmt.Errorf("failed to create permission: %w", err)
		}
		result = permission
		return nil
	})

	if err != nil {
		return nil, err
	}

	s.publisher.Publish(ctx, event.EventPermissionCreate, &event.EventPermissionCreateData{
		UID:         result.UID,
		Resource:    result.Resource,
		Action:      result.Action,
		Description: result.Description,
		CreatedAt:   result.CreatedAt,
	})
	return result, nil
}

func (s *permissionService) Get(ctx context.Context, uid string) (*model.Permission, error) {
	ids, err := s.resolvers.Permission().IDsByUIDs(ctx, []string{uid})
	if err != nil {
		s.observer.OnSignal(ctx, signal.SignalError, signal.SignalPermission{
			UID:       &uid,
			Operation: "get",
		}, err)
		return nil, domainerrors.ErrPermissionGetFailed
	}

	id, exists := ids[uid]
	if !exists {
		err := domainerrors.ErrPermissionNotFound
		s.observer.OnSignal(ctx, signal.SignalReject, signal.SignalPermission{
			UID:       &uid,
			Operation: "get",
		}, err)
		return nil, err
	}

	permission, err := s.repos.Permission().GetByID(ctx, id)
	if err != nil {
		s.observer.OnSignal(ctx, signal.SignalError, signal.SignalPermission{
			UID:       &uid,
			Operation: "get",
		}, err)
		return nil, domainerrors.ErrPermissionGetFailed
	}

	s.observer.OnSignal(ctx, signal.SignalSuccess, signal.SignalPermission{
		UID:       &uid,
		Resource:  &permission.Resource,
		Action:    &permission.Action,
		Operation: "get",
	}, nil)

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
	var permission *model.Permission

	err := s.uow.Do(ctx, func(r repository.RepositoryProvider) error {
		repo := r.Permission()

		var errUoW error
		permission, errUoW = repo.GetByUID(ctx, uid)
		if errUoW != nil {
			return fmt.Errorf("failed to get permission: %w", errUoW)
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
		return nil
	})

	if err != nil {
		return err
	}

	s.publisher.Publish(ctx, event.EventPermissionUpdate, &event.EventPermissionUpdateData{
		UID:         permission.UID,
		Resource:    permission.Resource,
		Action:      permission.Action,
		Description: permission.Description,
		UpdatedAt:   permission.UpdatedAt,
	})
	return nil
}

func (s *permissionService) Delete(ctx context.Context, uid string) error {
	var permission *model.Permission

	err := s.uow.Do(ctx, func(r repository.RepositoryProvider) error {
		repo := r.Permission()

		var errUoW error
		permission, errUoW = repo.GetByUID(ctx, uid)
		if errUoW != nil {
			return fmt.Errorf("failed to get permission: %w", errUoW)
		}

		if err := repo.Delete(ctx, permission.ID); err != nil {
			return fmt.Errorf("failed to delete permission: %w", err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	s.publisher.Publish(ctx, event.EventPermissionDelete, &event.EventPermissionDeleteData{
		UID: permission.UID,
	})
	return nil
}
