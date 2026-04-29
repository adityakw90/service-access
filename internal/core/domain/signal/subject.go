package signal

import "time"

type SignalSubject struct {
	SubjectID   *string
	SubjectType *string
	RoleUID     *string
	AssignedAt  *time.Time

	// operation context
	Operation string
}
