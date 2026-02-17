package param

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
