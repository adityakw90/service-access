package param

// PermissionCreateParam represents the parameters for creating a permission.
type PermissionCreateParam struct {
	Resource    string
	Action      string
	Description string
}

// PermissionUpdateParam represents the parameters for updating a permission.
type PermissionUpdateParam struct {
	Resource    *string
	Action      *string
	Description *string
}

// PermissionListFilterParam represents the parameters for filtering permissions.
type PermissionListFilterParam struct {
	UIDs     []string
	Resource *string
	Action   *string
	Query    *string
}

// PermissionListParam represents the parameters for listing permissions.
type PermissionListParam struct {
	Pagination *PaginationParam
	Filter     *PermissionListFilterParam
}
