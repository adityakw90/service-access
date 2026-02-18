package repository

import portrepository "github.com/adityakw90/service-access/internal/core/port/repository"

type repositories struct {
	db         dbExecutor
	permission portrepository.PermissionRepository
	group      portrepository.GroupRepository
	role       portrepository.RoleRepository
	subject    portrepository.SubjectRepository
}

func (r *repositories) Permission() portrepository.PermissionRepository {
	if r.permission == nil {
		r.permission = NewPermissionRepository(r.db)
	}
	return r.permission
}

func (r *repositories) Group() portrepository.GroupRepository {
	if r.group == nil {
		r.group = NewGroupRepository(r.db)
	}
	return r.group
}

func (r *repositories) Role() portrepository.RoleRepository {
	if r.role == nil {
		r.role = NewRoleRepository(r.db)
	}
	return r.role
}

func (r *repositories) Subject() portrepository.SubjectRepository {
	if r.subject == nil {
		r.subject = NewSubjectRepository(r.db)
	}
	return r.subject
}
