package repository

import (
	portrepository "github.com/adityakw90/service-access/internal/core/port/repository"
)

// repositoryProvider provides non-transactional access to repositories.
type repositoryProvider struct {
	db PostgrePool
}

// NewRepositoryProvider creates a new RepositoryProvider.
func NewRepositoryProvider(db PostgrePool) portrepository.RepositoryProvider {
	return &repositoryProvider{db: db}
}

func (p *repositoryProvider) Permission() portrepository.PermissionRepository {
	return NewPermissionRepository(p.db)
}

func (p *repositoryProvider) Group() portrepository.GroupRepository {
	return NewGroupRepository(p.db)
}

func (p *repositoryProvider) Role() portrepository.RoleRepository {
	return NewRoleRepository(p.db)
}

func (p *repositoryProvider) Subject() portrepository.SubjectRepository {
	return NewSubjectRepository(p.db)
}
