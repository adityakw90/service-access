package param

// GroupOrderBy represents allowed OrderBy column values for Group.
type GroupOrderBy string

const (
	OrderByGroupID          GroupOrderBy = "id"
	OrderByGroupUID         GroupOrderBy = "uid"
	OrderByGroupName        GroupOrderBy = "name"
	OrderByGroupDescription GroupOrderBy = "description"
	OrderByGroupCreatedAt   GroupOrderBy = "created_at"
	OrderByGroupUpdatedAt   GroupOrderBy = "updated_at"
)

// GroupPermissionOrderBy represents allowed OrderBy column values for GroupPermission.
type GroupPermissionOrderBy string

const (
	OrderByGroupPermissionGroupID      GroupPermissionOrderBy = "group_id"
	OrderByGroupPermissionPermissionID GroupPermissionOrderBy = "permission_id"
	OrderByGroupPermissionCreatedAt    GroupPermissionOrderBy = "created_at"
)

// GroupCreateParam represents the parameters for creating a group.
type GroupCreateParam struct {
	Name        string
	Description string
}

// GroupUpdateParam represents the parameters for updating a group.
type GroupUpdateParam struct {
	Name        *string
	Description *string
}

// GroupListFilterParam represents the parameters for filtering groups.
type GroupListFilterParam struct {
	IDs   []int64
	UIDs  []string
	Name  *string
	Query *string
}

// GroupListParam represents the parameters for listing groups.
type GroupListParam struct {
	Pagination *PaginationParam
	Filter     *GroupListFilterParam
}

type GroupPermissionListFilterParam struct {
	IDs            []int64
	UIDs           []string
	PermissionIDs  []int64
	PermissionUIDs []string
	Resource       *string
	Action         *string
	Query          *string
}

// GroupPermissionListParam represents the parameters for listing group permissions.
type GroupPermissionListParam struct {
	Pagination *PaginationParam
	Filter     *GroupPermissionListFilterParam
}

type GroupPermissionMapGroupIDPermissionID struct {
	GroupID      int64
	PermissionID int64
}
