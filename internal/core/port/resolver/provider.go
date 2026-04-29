package resolver

type ResolverProvider interface {
	Permission() PermissionResolver
	Group() GroupResolver
	Role() RoleResolver
	GroupPermission() GroupPermissionResolver
}
