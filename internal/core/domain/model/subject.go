package model

import "time"

// SubjectRole represents the assignment of a role to a subject (user, service account, etc.).
type SubjectRole struct {
	SubjectID   string
	SubjectType string
	RoleID      int64
	RoleUID     string
	AssignedAt  time.Time
}

type SubjectRoles struct {
	Items []SubjectRole
	Meta  Meta
}
