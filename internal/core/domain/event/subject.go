package event

type EventSubjectAssignData struct {
	SubjectID   string
	SubjectType string
	RoleID      int64
	RoleUID     string
}

type EventSubjectRevokeData struct {
	SubjectID   string
	SubjectType string
	RoleID      int64
	RoleUID     string
}
