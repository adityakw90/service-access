package event

type EventAccessCheckData struct {
	SubjectId   string
	SubjectType string
	Resource    string
	Action      string
	Reason      string
}

// NewEventAccessChecked creates a new access checked event.
func NewEventAccessChecked(subjectId, subjectType, resource, action, reason string, allowed bool) Event {
	return newEvent(EventAccessCheck, EventAccessCheckData{
		SubjectId:   subjectId,
		SubjectType: subjectType,
		Resource:    resource,
		Action:      action,
		Reason:      reason,
	})
}
