package repository

import (
	"sync"

	portrepository "github.com/adityakw90/service-access/internal/core/port/repository"
)

type repositories struct {
	db         dbExecutor
	permission portrepository.PermissionRepository
	permissionOnce sync.Once
	group      portrepository.GroupRepository
	groupOnce sync.Once
	role       portrepository.RoleRepository
	roleOnce sync.Once
	subject    portrepository.SubjectRepository
	subjectOnce sync.Once
}

func (r *repositories) Permission() portrepository.PermissionRepository {
	r.permissionOnce.Do(func() {
		r.permission = NewPermissionRepository(r.db)
	})
	return r.permission
}

func (r *repositories) Group() portrepository.GroupRepository {
	r.groupOnce.Do(func() {
		r.group = NewGroupRepository(r.db)
	})
	return r.group
}

func (r *repositories) Role() portrepository.RoleRepository {
	r.roleOnce.Do(func() {
		r.role = NewRoleRepository(r.db)
	})
	return r.role
}

func (r *repositories) Subject() portrepository.SubjectRepository {
	r.subjectOnce.Do(func() {
		r.subject = NewSubjectRepository(r.db)
	})
	return r.subject
}
