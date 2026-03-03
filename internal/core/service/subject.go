package service

import (
	"context"
	"fmt"
	"time"

	"github.com/adityakw90/service-access/internal/core/domain/event"
	"github.com/adityakw90/service-access/internal/core/domain/model"
	"github.com/adityakw90/service-access/internal/core/domain/param"
	portEvent "github.com/adityakw90/service-access/internal/core/port/event"
	"github.com/adityakw90/service-access/internal/core/port/repository"
	"github.com/adityakw90/service-access/internal/core/port/service"
)

type subjectService struct {
	uow       repository.UnitOfWork
	repos     repository.RepositoryProvider
	publisher portEvent.EventPublisher
}

func NewSubjectService(uow repository.UnitOfWork, repos repository.RepositoryProvider, publisher portEvent.EventPublisher) service.SubjectService {
	return &subjectService{
		uow:       uow,
		repos:     repos,
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
	var role *model.Role
	var subjectRole *model.SubjectRole

	err := s.uow.Do(ctx, func(r repository.RepositoryProvider) error {
		var errUoW error
		role, errUoW = r.Role().GetByUID(ctx, roleUID)
		if errUoW != nil {
			return fmt.Errorf("failed to get role: %w", errUoW)
		}

		// Create subject-role assignment
		subjectRole = &model.SubjectRole{
			SubjectID:   subjectID,
			SubjectType: subjectType,
			RoleID:      role.ID,
		}

		if err := r.Subject().Create(ctx, subjectRole); err != nil {
			return fmt.Errorf("failed to assign role: %w", err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	s.publisher.Publish(ctx, event.EventSubjectAssign, &event.EventSubjectAssignData{
		SubjectID:   subjectID,
		SubjectType: subjectType,
		RoleUID:     role.UID,
		AssignedAt:  subjectRole.AssignedAt,
	})
	return nil
}

func (s *subjectService) Revoke(ctx context.Context, subjectID string, subjectType string, roleUID string) error {
	var role *model.Role

	err := s.uow.Do(ctx, func(r repository.RepositoryProvider) error {
		var errUoW error
		role, errUoW = r.Role().GetByUID(ctx, roleUID)
		if errUoW != nil {
			return fmt.Errorf("failed to get role: %w", errUoW)
		}

		// Delete the subject-role assignment
		if err := r.Subject().Delete(ctx, subjectID, subjectType, role.ID); err != nil {
			return fmt.Errorf("failed to revoke role: %w", err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	s.publisher.Publish(ctx, event.EventSubjectRevoke, &event.EventSubjectRevokeData{
		SubjectID:   subjectID,
		SubjectType: subjectType,
		RoleUID:     role.UID,
		RevokedAt:   time.Now(),
	})
	return nil
}
