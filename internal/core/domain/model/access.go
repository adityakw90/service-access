package model

import "time"

type Access struct {
	SubjectID     string
	SubjectType   string
	Resource      string
	Action        string
	PermissionID  *int64
	PermissionUID *string
	RoleID        *int64
	RoleUID       *string
	GroupID       *int64
	GroupUID      *string
	AssignedAt    *time.Time
}
