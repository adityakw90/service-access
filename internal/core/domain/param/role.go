package param

// RoleOrderBy represents allowed OrderBy column values for Role.
type RoleOrderBy string

const (
	OrderByRoleID          RoleOrderBy = "id"
	OrderByRoleUID         RoleOrderBy = "uid"
	OrderByRoleGroupID     RoleOrderBy = "group_id"
	OrderByRoleName        RoleOrderBy = "name"
	OrderByRoleDescription RoleOrderBy = "description"
	OrderByRoleCreatedAt   RoleOrderBy = "created_at"
	OrderByRoleUpdatedAt   RoleOrderBy = "updated_at"
)

// RolePermissionOrderBy represents allowed OrderBy column values for RolePermission.
type RolePermissionOrderBy string

const (
	OrderByRolePermissionRoleID            RolePermissionOrderBy = "role_id"
	OrderByRolePermissionGroupPermissionID RolePermissionOrderBy = "group_permission_id"
	OrderByRolePermissionCreatedAt         RolePermissionOrderBy = "created_at"
)

// RoleCreateParam represents the parameters for creating a role.
type RoleCreateParam struct {
	GroupID     int64
	GroupUID    string
	Name        string
	Description string
}

// RoleUpdateParam represents the parameters for updating a role.
type RoleUpdateParam struct {
	Name        *string
	Description *string
}

// ListRolesParam represents the parameters for listing roles.
type RoleListFilterParam struct {
	IDs      []int64
	UIDs     []string
	GroupID  *int64
	GroupUID *string
	Name     *string
	Query    *string
}

// RoleListParam represents the parameters for listing roles.
type RoleListParam struct {
	Pagination *PaginationParam
	Filter     *RoleListFilterParam
}

// RolePermissionListFilterParam represents the parameters for filtering role permissions.
type RolePermissionListFilterParam struct {
	IDs            []int64
	UIDs           []string
	PermissionIDs  []int64
	PermissionUIDs []string
	Resource       *string
	Action         *string
	Query          *string
}

// RolePermissionListParam represents the parameters for listing role permissions.
type RolePermissionListParam struct {
	Pagination *PaginationParam
	Filter     *RolePermissionListFilterParam
}
