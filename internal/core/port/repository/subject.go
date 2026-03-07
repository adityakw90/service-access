package repository

import (
	"context"

	"github.com/adityakw90/service-access/internal/core/domain/model"
	"github.com/adityakw90/service-access/internal/core/domain/param"
)

// SubjectRepository defines the interface for subject repository.
type SubjectRepository interface {
	// Manage
	Create(ctx context.Context, subject *model.SubjectRole) error                         // Create creates a subject role.
	Update(ctx context.Context, subject *model.SubjectRole) error                         // Update updates a subject role.
	Delete(ctx context.Context, subjectID string, subjectType string, roleID int64) error // Delete deletes a subject role.

	// Retrieve
	List(ctx context.Context, pagination *param.PaginationParam, filter *param.SubjectListFilterParam) (model.SubjectRoles, error) // List returns the list of subject roles.
	GetRoles(ctx context.Context, subjectID string, subjectType string) ([]model.SubjectRole, error)                               // GetRoles returns roles for a subject.
	GetAllRoles(ctx context.Context, subjectID string, subjectType string) ([]model.Role, error)                                  // GetAllRoles returns all roles assigned to a subject (direct assignments only).
}
