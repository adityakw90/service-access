package service

import (
	"context"

	"github.com/adityakw90/service-access/internal/core/domain/model"
	"github.com/adityakw90/service-access/internal/core/domain/param"
)

type SubjectService interface {
	// retrieve
	List(ctx context.Context, pagination *param.PaginationParam, filter *param.SubjectListFilterParam) (*model.SubjectRoles, error)

	// manage
	Assign(ctx context.Context, subjectID string, subjectType string, roleUID string) error // Assign role to subject
	Revoke(ctx context.Context, subjectID string, subjectType string, roleUID string) error // Revoke role from subject

	// aggregation methods
	GetRoles(ctx context.Context, subjectID string, subjectType string) ([]model.Role, error)
	GetGroups(ctx context.Context, subjectID string, subjectType string) ([]model.Group, error)
	GetPermissions(ctx context.Context, subjectID string, subjectType string) ([]model.Permission, error)
	GetFullProfile(ctx context.Context, subjectID string, subjectType string) (*model.SubjectProfile, error)
}
