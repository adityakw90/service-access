package repository

type RepositoryProvider interface {
	Permission() PermissionRepository
	Group() GroupRepository
	Role() RoleRepository
	Subject() SubjectRepository
}
