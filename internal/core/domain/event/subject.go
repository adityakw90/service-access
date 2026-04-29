package event

import (
	"time"

	"github.com/adityakw90/service-access/internal/core/domain/model"
)

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

// NewEventSubjectAssigned creates a new subject assigned event.
func NewEventSubjectAssigned(subjectRole *model.SubjectRole, role *model.Role) Event {
	return newEvent(EventSubjectAssign, EventSubjectAssignData{
		SubjectID:   subjectRole.SubjectID,
		SubjectType: subjectRole.SubjectType,
		RoleUID:     role.UID,
		AssignedAt:  subjectRole.AssignedAt,
	})
}

// NewEventSubjectRevoked creates a new subject revoked event.
func NewEventSubjectRevoked(subjectID, subjectType, roleUID string) Event {
	return newEvent(EventSubjectRevoke, EventSubjectRevokeData{
		SubjectID:   subjectID,
		SubjectType: subjectType,
		RoleUID:     roleUID,
		RevokedAt:   time.Now(),
	})
}
