package event

import (
	"time"

	"github.com/adityakw90/service-access/internal/core/domain/model"
)

type EventRoleCreateData struct {
	UID         string
	GroupUID    string
	Name        string
	Description string
	CreatedAt   time.Time
}

type EventRoleUpdateData struct {
	UID         string
	Name        string
	Description string
	UpdatedAt   time.Time
}

type EventRoleDeleteData struct {
	UID string
}

type EventRoleAssignPermissionData struct {
	RoleUID            string
	GroupPermissionUID string
}

type EventRoleRevokePermissionData struct {
	RoleUID            string
	GroupPermissionUID string
}

type EventRoleUpdatePermissionData struct {
	RoleUID string
	UIDs    []string
}

// NewEventRoleCreated creates a new role created event.
func NewEventRoleCreated(role *model.Role) Event {
	return newEvent(EventRoleCreate, EventRoleCreateData{
		UID:         role.UID,
		GroupUID:    role.GroupUID,
		Name:        role.Name,
		Description: role.Description,
		CreatedAt:   role.CreatedAt,
	})
}

// NewEventRoleUpdated creates a new role updated event.
func NewEventRoleUpdated(role *model.Role) Event {
	return newEvent(EventRoleUpdate, EventRoleUpdateData{
		UID:         role.UID,
		Name:        role.Name,
		Description: role.Description,
		UpdatedAt:   role.UpdatedAt,
	})
}

// NewEventRoleDeleted creates a new role deleted event.
func NewEventRoleDeleted(role *model.Role) Event {
	return newEvent(EventRoleDelete, EventRoleDeleteData{
		UID: role.UID,
	})
}

// NewEventRolePermissionAssigned creates a new role permission assigned event.
func NewEventRolePermissionAssigned(role *model.Role, groupPermission *model.GroupPermission) Event {
	return newEvent(EventRoleAssignPermission, EventRoleAssignPermissionData{
		RoleUID:            role.UID,
		GroupPermissionUID: groupPermission.UID,
	})
}

// NewEventRolePermissionRevoked creates a new role permission revoked event.
func NewEventRolePermissionRevoked(role *model.Role, groupPermission *model.GroupPermission) Event {
	return newEvent(EventRoleRevokePermission, EventRoleRevokePermissionData{
		RoleUID:            role.UID,
		GroupPermissionUID: groupPermission.UID,
	})
}

// NewEventRolePermissionsUpdated creates a new role permissions updated event.
func NewEventRolePermissionsUpdated(role *model.Role) Event {
	return newEvent(EventRoleUpdatePermission, EventRoleUpdatePermissionData{
		RoleUID: role.UID,
	})
}
