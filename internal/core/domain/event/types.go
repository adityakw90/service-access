package event

// Event represents a domain event.
type Event interface {
	// Type returns the event type.
	Type() EventType

	// Data returns the event data.
	Data() any
}

// Publisher publishes domain events.
// This is a simplified interface used by the service layer.
// type Publisher interface {
// 	Publish(ctx context.Context, events []Event) error
// }
// Use interface from port/event

// EventType defines the type of access domain event.
type EventType string

const (
	// Access Events
	EventAccessCheck EventType = "access.check"

	// Subject Events
	EventSubjectAssign EventType = "subject.assign"
	EventSubjectRevoke EventType = "subject.revoke"

	// Role Events
	EventRoleCreate           EventType = "role.create"
	EventRoleUpdate           EventType = "role.update"
	EventRoleDelete           EventType = "role.delete"
	EventRoleUpdatePermission EventType = "role.update_permission"
	EventRoleAssignPermission EventType = "role.assign_permission"
	EventRoleRevokePermission EventType = "role.revoke_permission"

	// Group Events
	EventGroupCreate           EventType = "group.create"
	EventGroupUpdate           EventType = "group.update"
	EventGroupDelete           EventType = "group.delete"
	EventGroupUpdatePermission EventType = "group.update_permission"
	EventGroupAssignPermission EventType = "group.assign_permission"
	EventGroupRevokePermission EventType = "group.revoke_permission"

	// Permission Events
	EventPermissionCreate EventType = "permission.create"
	EventPermissionUpdate EventType = "permission.update"
	EventPermissionDelete EventType = "permission.delete"
)

// baseEvent implements the Event interface.
type baseEvent struct {
	eventType EventType
	data      any
}

// Type returns the event type.
func (e *baseEvent) Type() EventType {
	return e.eventType
}

// Data returns the event data.
func (e *baseEvent) Data() any {
	return e.data
}

// newEvent creates a new base event.
func newEvent(eventType EventType, data any) Event {
	return &baseEvent{
		eventType: eventType,
		data:      data,
	}
}
