package service

import (
	"context"
	"fmt"

	domainSignal "github.com/adityakw90/service-access/internal/core/domain/signal"
	portEvent "github.com/adityakw90/service-access/internal/core/port/event"
	portExecutor "github.com/adityakw90/service-access/internal/core/port/executor"
	portObserver "github.com/adityakw90/service-access/internal/core/port/observer"
	"github.com/adityakw90/service-access/internal/core/port/repository"
	"github.com/adityakw90/service-access/internal/core/port/service"
)

type accessService struct {
	repos          repository.RepositoryProvider
	eventPublisher portEvent.EventPublisher
	executor       portExecutor.Executor
	accessObserver portObserver.ServiceObserver[domainSignal.SignalAccessCheck]
}

func NewAccessService(
	repos repository.RepositoryProvider,
	eventPublisher portEvent.EventPublisher,
	executor portExecutor.Executor,
	accessObserver portObserver.ServiceObserver[domainSignal.SignalAccessCheck],
) service.AccessService {
	return &accessService{
		repos:          repos,
		eventPublisher: eventPublisher,
		executor:       executor,
		accessObserver: accessObserver,
	}
}

func (s *accessService) Check(ctx context.Context, subjectID string, subjectType string, resource string, action string) (allowed bool, reason string, err error) {
	// Get all roles for subject
	subjectRoles, err := s.repos.Subject().GetRoles(ctx, subjectID, subjectType)
	if err != nil {
		return false, "", fmt.Errorf("failed to get subject roles: %w", err)
	}

	// Collect all permissions from all roles
	permissionIDs := make([]int64, 0)
	for _, sr := range subjectRoles {
		rolePermissions, err := s.repos.Role().ListPermission(ctx, sr.RoleID, nil, nil)
		if err != nil {
			continue
		}
		for _, rp := range rolePermissions.Items {
			// Get the group permission to find the actual permission ID
			groupPerm, err := s.repos.Group().GetPermissionByID(ctx, rp.GroupPermissionID)
			if err != nil {
				continue
			}
			permissionIDs = append(permissionIDs, groupPerm.PermissionID)
		}
	}

	// Check if any permission grants access
	for _, pid := range permissionIDs {
		permission, err := s.repos.Permission().GetByID(ctx, pid)
		if err != nil {
			continue
		}
		if permission.Resource == resource && permission.Action == action {
			return true, "permission granted", nil
		}
	}

	return false, "no matching permission found", nil
}
