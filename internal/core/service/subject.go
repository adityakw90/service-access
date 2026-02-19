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

type subjectService struct {
	uow      repository.UnitOfWork
	repos    repository.RepositoryProvider
	publisher portEvent.EventPublisher
}

func NewSubjectService(uow repository.UnitOfWork, repos repository.RepositoryProvider, publisher portEvent.EventPublisher) service.SubjectService {
	return &subjectService{
		uow:      uow,
		repos:    repos,
		publisher: publisher,
	}
}

func (s *subjectService) List(ctx context.Context, pagination *param.PaginationParam, filter *param.SubjectListFilterParam) (*model.SubjectRoles, error) {
	subjects, err := s.repos.Subject().List(ctx, pagination, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list subjects: %w", err)
	}
	return &subjects, nil
}

func (s *subjectService) Assign(ctx context.Context, subjectID string, subjectType string, roleUID string) error {
	var events []event.Event

	err := s.uow.Do(ctx, func(r repository.Repositories) error {
		role, err := r.Role().GetByUID(ctx, roleUID)
		if err != nil {
			return fmt.Errorf("failed to get role: %w", err)
		}

		// Create subject-role assignment
		subjectRole := &model.SubjectRole{
			SubjectID:   subjectID,
			SubjectType: subjectType,
			RoleID:      role.ID,
		}

		if err := r.Subject().Create(ctx, subjectRole); err != nil {
			return fmt.Errorf("failed to assign role: %w", err)
		}

		events = []event.Event{event.NewEventSubjectAssigned(subjectRole, role)}
		return nil
	})

	if err != nil {
		return err
	}

	s.publisher.Publish(ctx, events)
	return nil
}

func (s *subjectService) Revoke(ctx context.Context, subjectID string, subjectType string, roleUID string) error {
	var events []event.Event

	err := s.uow.Do(ctx, func(r repository.Repositories) error {
		role, err := r.Role().GetByUID(ctx, roleUID)
		if err != nil {
			return fmt.Errorf("failed to get role: %w", err)
		}

		// Delete the subject-role assignment
		err = r.Subject().Delete(ctx, role.ID, subjectType, role.ID)
		if err != nil {
			return fmt.Errorf("failed to revoke role: %w", err)
		}

		events = []event.Event{event.NewEventSubjectRevoked(subjectID, subjectType, role.UID)}
		return nil
	})

	if err != nil {
		return err
	}

	s.publisher.Publish(ctx, events)
	return nil
}
