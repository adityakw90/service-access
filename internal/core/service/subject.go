package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/adityakw90/service-access/internal/core/domain/event"
	"github.com/adityakw90/service-access/internal/core/domain/model"
	"github.com/adityakw90/service-access/internal/core/domain/param"
	"github.com/adityakw90/service-access/internal/core/domain/signal"
	portEvent "github.com/adityakw90/service-access/internal/core/port/event"
	"github.com/adityakw90/service-access/internal/core/port/observer"
	"github.com/adityakw90/service-access/internal/core/port/repository"
	"github.com/adityakw90/service-access/internal/core/port/service"
	"golang.org/x/sync/errgroup"
)

type subjectService struct {
	uow             repository.UnitOfWork
	repos           repository.RepositoryProvider
	publisher       portEvent.EventPublisher
	subjectObserver observer.ServiceObserver[signal.SignalSubject]
}

func NewSubjectService(
	uow repository.UnitOfWork,
	repos repository.RepositoryProvider,
	publisher portEvent.EventPublisher,
	subjectObserver observer.ServiceObserver[signal.SignalSubject],
) service.SubjectService {
	return &subjectService{
		uow:             uow,
		repos:           repos,
		publisher:       publisher,
		subjectObserver: subjectObserver,
	}
}

func (s *subjectService) List(ctx context.Context, pagination *param.PaginationParam, filter *param.SubjectListFilterParam) (*model.SubjectRoles, error) {
	subjects, err := s.repos.Subject().List(ctx, pagination, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list subjects: %w", err)
	}

	s.subjectObserver.OnSignal(ctx, signal.SignalSuccess, signal.SignalSubject{
		Operation: "list",
	}, nil)
	return &subjects, nil
}

func (s *subjectService) Assign(ctx context.Context, subjectID string, subjectType string, roleUID string) error {
	var role *model.Role
	var subjectRole *model.SubjectRole

	err := s.uow.Do(ctx, func(r repository.RepositoryProvider) error {
		var errUoW error
		role, errUoW = r.Role().GetByUID(ctx, roleUID)
		if errUoW != nil {
			s.subjectObserver.OnSignal(ctx, signal.SignalError, signal.SignalSubject{
				SubjectID:   &subjectID,
				SubjectType: &subjectType,
				RoleUID:     &roleUID,
				Operation:   "assign",
			}, errUoW)
			return fmt.Errorf("failed to get role: %w", errUoW)
		}

		// Create subject-role assignment
		subjectRole = &model.SubjectRole{
			SubjectID:   subjectID,
			SubjectType: subjectType,
			RoleID:      role.ID,
		}

		if err := r.Subject().Create(ctx, subjectRole); err != nil {
			s.subjectObserver.OnSignal(ctx, signal.SignalError, signal.SignalSubject{
				SubjectID:   &subjectID,
				SubjectType: &subjectType,
				RoleUID:     &roleUID,
				Operation:   "assign",
			}, err)
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
	s.subjectObserver.OnSignal(ctx, signal.SignalSuccess, signal.SignalSubject{
		SubjectID:   &subjectID,
		SubjectType: &subjectType,
		RoleUID:     &role.UID,
		Operation:   "assign",
	}, nil)
	return nil
}

func (s *subjectService) Revoke(ctx context.Context, subjectID string, subjectType string, roleUID string) error {
	var role *model.Role

	err := s.uow.Do(ctx, func(r repository.RepositoryProvider) error {
		var errUoW error
		role, errUoW = r.Role().GetByUID(ctx, roleUID)
		if errUoW != nil {
			s.subjectObserver.OnSignal(ctx, signal.SignalError, signal.SignalSubject{
				SubjectID:   &subjectID,
				SubjectType: &subjectType,
				RoleUID:     &roleUID,
				Operation:   "revoke",
			}, errUoW)
			return fmt.Errorf("failed to get role: %w", errUoW)
		}

		// Delete the subject-role assignment
		if err := r.Subject().Delete(ctx, subjectID, subjectType, role.ID); err != nil {
			s.subjectObserver.OnSignal(ctx, signal.SignalError, signal.SignalSubject{
				SubjectID:   &subjectID,
				SubjectType: &subjectType,
				RoleUID:     &roleUID,
				Operation:   "revoke",
			}, err)
			return fmt.Errorf("failed to revoke role: %w", err)
		}
		return nil
	})

	if err != nil {
		s.subjectObserver.OnSignal(ctx, signal.SignalError, signal.SignalSubject{
			SubjectID:   &subjectID,
			SubjectType: &subjectType,
			RoleUID:     &roleUID,
			Operation:   "revoke",
		}, err)
		return err
	}

	s.publisher.Publish(ctx, event.EventSubjectRevoke, &event.EventSubjectRevokeData{
		SubjectID:   subjectID,
		SubjectType: subjectType,
		RoleUID:     role.UID,
		RevokedAt:   time.Now(),
	})

	s.subjectObserver.OnSignal(ctx, signal.SignalSuccess, signal.SignalSubject{
		SubjectID:   &subjectID,
		SubjectType: &subjectType,
		RoleUID:     &role.UID,
		Operation:   "revoke",
	}, nil)

	return nil
}

// Validation errors
var (
	ErrInvalidSubjectID = errors.New("subject_id is required")
	ErrInvalidSubjectType = errors.New("subject_type is required")
)

func validateSubjectParams(subjectID, subjectType string) error {
	if subjectID == "" {
		return ErrInvalidSubjectID
	}
	if subjectType == "" {
		return ErrInvalidSubjectType
	}
	return nil
}

// GetRoles returns all roles assigned to the subject.
func (s *subjectService) GetRoles(ctx context.Context, subjectID string, subjectType string) ([]model.Role, error) {
	if err := validateSubjectParams(subjectID, subjectType); err != nil {
		return nil, err
	}

	roles, err := s.repos.Subject().GetAllRoles(ctx, subjectID, subjectType)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles for subject %s: %w", subjectID, err)
	}
	return roles, nil
}

// GetGroups returns all unique groups from the subject's roles.
func (s *subjectService) GetGroups(ctx context.Context, subjectID string, subjectType string) ([]model.Group, error) {
	if err := validateSubjectParams(subjectID, subjectType); err != nil {
		return nil, err
	}

	groups, err := s.repos.Subject().GetAllGroups(ctx, subjectID, subjectType)
	if err != nil {
		return nil, fmt.Errorf("failed to get groups for subject %s: %w", subjectID, err)
	}
	return groups, nil
}

// GetPermissions returns all unique permissions from all of the subject's roles.
func (s *subjectService) GetPermissions(ctx context.Context, subjectID string, subjectType string) ([]model.Permission, error) {
	if err := validateSubjectParams(subjectID, subjectType); err != nil {
		return nil, err
	}

	perms, err := s.repos.Subject().GetAllPermissions(ctx, subjectID, subjectType)
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions for subject %s: %w", subjectID, err)
	}
	return perms, nil
}

// GetFullProfile returns complete aggregation (groups, roles, permissions) in parallel.
func (s *subjectService) GetFullProfile(ctx context.Context, subjectID string, subjectType string) (*model.SubjectProfile, error) {
	if err := validateSubjectParams(subjectID, subjectType); err != nil {
		return nil, err
	}

	profile := &model.SubjectProfile{}
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(10) // limit concurrency

	// Fetch groups
	g.Go(func() error {
		groups, err := s.repos.Subject().GetAllGroups(ctx, subjectID, subjectType)
		if err != nil {
			return fmt.Errorf("failed to get groups: %w", err)
		}
		profile.Groups = groups
		return nil
	})

	// Fetch roles
	g.Go(func() error {
		roles, err := s.repos.Subject().GetAllRoles(ctx, subjectID, subjectType)
		if err != nil {
			return fmt.Errorf("failed to get roles: %w", err)
		}
		profile.Roles = roles
		return nil
	})

	// Fetch permissions
	g.Go(func() error {
		perms, err := s.repos.Subject().GetAllPermissions(ctx, subjectID, subjectType)
		if err != nil {
			return fmt.Errorf("failed to get permissions: %w", err)
		}
		profile.Permissions = perms
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return profile, nil
}
