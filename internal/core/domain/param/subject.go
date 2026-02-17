package param

type SubjectListFilterParam struct {
	SubjectIDs   *string
	SubjectTypes *string
	RoleUIDs     *string
	Query        *string
}

type SubjectListParam struct {
	Pagination *PaginationParam
	Filter     *SubjectListFilterParam
}
