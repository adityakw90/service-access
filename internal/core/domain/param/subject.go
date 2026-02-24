package param

type SubjectAssignRoleParam struct {
	SubjectID   string
	SubjectType string
	RoleUID     string
}

type SubjectRevokeRoleParam struct {
	SubjectID   string
	SubjectType string
	RoleUID     string
}

type SubjectListFilterParam struct {
	SubjectID   *string
	SubjectType *string
	RoleID      *int64
	RoleUID     *string
	Query       *string
}

type SubjectListParam struct {
	Pagination *PaginationParam
	Filter     *SubjectListFilterParam
}
