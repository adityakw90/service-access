package event

// EventType defines the type of access domain event.
type EventType string

const (
	// Access Events
	EventCheckAccess       EventType = "access.check"
	EventCheckAccessFailed EventType = "access.check_failed"
)
