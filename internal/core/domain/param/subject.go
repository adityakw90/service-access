package param

// SubjectOrderBy represents allowed OrderBy column values for Subject.
type SubjectOrderBy string

const (
	OrderBySubjectID         SubjectOrderBy = "subject_id"
	OrderBySubjectType       SubjectOrderBy = "subject_type"
	OrderBySubjectRoleID     SubjectOrderBy = "role_id"
	OrderBySubjectAssignedAt SubjectOrderBy = "assigned_at"
)

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

type SubjectGetParam struct {
	SubjectID   string
	SubjectType string
}
