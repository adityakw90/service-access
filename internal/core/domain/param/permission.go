package param

// PermissionOrderBy represents allowed OrderBy column values for Permission.
type PermissionOrderBy string

const (
	OrderByPermissionID          PermissionOrderBy = "id"
	OrderByPermissionUID         PermissionOrderBy = "uid"
	OrderByPermissionResource    PermissionOrderBy = "resource"
	OrderByPermissionAction      PermissionOrderBy = "action"
	OrderByPermissionDescription PermissionOrderBy = "description"
	OrderByPermissionCreatedAt   PermissionOrderBy = "created_at"
	OrderByPermissionUpdatedAt   PermissionOrderBy = "updated_at"
)

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
	IDs      []int64
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

type PermissionMapResourceAction struct {
	Resource string
	Action   string
}
