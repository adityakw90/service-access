package event

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
