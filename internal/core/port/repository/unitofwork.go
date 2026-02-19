package repository

import (
	"context"
)

type Repositories interface {
	Permission() PermissionRepository
	Group() GroupRepository
	Role() RoleRepository
	Subject() SubjectRepository
}

type RepositoryProvider interface {
	Permission() PermissionRepository
	Group() GroupRepository
	Role() RoleRepository
	Subject() SubjectRepository
}

type UnitOfWork interface {
	Do(ctx context.Context, fn func(r Repositories) error) error
}
