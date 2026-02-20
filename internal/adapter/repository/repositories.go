package repository

import (
	"sync"

	portrepository "github.com/adityakw90/service-access/internal/core/port/repository"
)

type repositoryProvider struct {
	db             dbExecutor
	permission     portrepository.PermissionRepository
	permissionOnce sync.Once
	group          portrepository.GroupRepository
	groupOnce      sync.Once
	role           portrepository.RoleRepository
	roleOnce       sync.Once
	subject        portrepository.SubjectRepository
	subjectOnce    sync.Once
}

// NewRepositoryProvider creates a new RepositoryProvider.
func NewRepositoryProvider(db dbExecutor) portrepository.RepositoryProvider {
	return &repositoryProvider{db: db}
}

func (r *repositoryProvider) Permission() portrepository.PermissionRepository {
	r.permissionOnce.Do(func() {
		r.permission = NewPermissionRepository(r.db)
	})
	return r.permission
}

func (r *repositoryProvider) Group() portrepository.GroupRepository {
	r.groupOnce.Do(func() {
		r.group = NewGroupRepository(r.db)
	})
	return r.group
}

func (r *repositoryProvider) Role() portrepository.RoleRepository {
	r.roleOnce.Do(func() {
		r.role = NewRoleRepository(r.db)
	})
	return r.role
}

func (r *repositoryProvider) Subject() portrepository.SubjectRepository {
	r.subjectOnce.Do(func() {
		r.subject = NewSubjectRepository(r.db)
	})
	return r.subject
}
