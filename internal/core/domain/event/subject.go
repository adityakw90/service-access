package event

import "time"

type EventSubjectAssignData struct {
	SubjectID   string
	SubjectType string
	RoleUID     string
	AssignedAt  time.Time
}

type EventSubjectRevokeData struct {
	SubjectID   string
	SubjectType string
	RoleUID     string
	RevokedAt   time.Time
}
