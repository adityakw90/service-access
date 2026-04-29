package event

import (
	"time"

	"github.com/adityakw90/service-access/internal/core/domain/model"
)

type EventGroupCreateData struct {
	UID         string
	Name        string
	Description string
	CreatedAt   time.Time
}

type EventGroupUpdateData struct {
	UID         string
	Name        string
	Description string
	UpdatedAt   time.Time
}

type EventGroupDeleteData struct {
	UID string
}

type EventGroupAssignPermissionData struct {
	GroupUID      string
	PermissionUID string
}

type EventGroupRevokePermissionData struct {
	GroupUID      string
	PermissionUID string
}

type EventGroupUpdatePermissionData struct {
	GroupUID string
	UIDs     []string
}

// NewEventGroupCreated creates a new group created event.
func NewEventGroupCreated(group *model.Group) Event {
	return newEvent(EventGroupCreate, EventGroupCreateData{
		UID:         group.UID,
		Name:        group.Name,
		Description: group.Description,
		CreatedAt:   group.CreatedAt,
	})
}

// NewEventGroupUpdated creates a new group updated event.
func NewEventGroupUpdated(group *model.Group) Event {
	return newEvent(EventGroupUpdate, EventGroupUpdateData{
		UID:         group.UID,
		Name:        group.Name,
		Description: group.Description,
		UpdatedAt:   group.UpdatedAt,
	})
}

// NewEventGroupDeleted creates a new group deleted event.
func NewEventGroupDeleted(group *model.Group) Event {
	return newEvent(EventGroupDelete, EventGroupDeleteData{
		UID: group.UID,
	})
}

// NewEventPermissionAssignedToGroup creates a new permission assigned to group event.
func NewEventPermissionAssignedToGroup(group *model.Group, permission *model.Permission) Event {
	return newEvent(EventGroupAssignPermission, EventGroupAssignPermissionData{
		GroupUID:      group.UID,
		PermissionUID: permission.UID,
	})
}

// NewEventPermissionRevokedFromGroup creates a new permission revoked from group event.
func NewEventPermissionRevokedFromGroup(group *model.Group, permission *model.Permission) Event {
	return newEvent(EventGroupRevokePermission, EventGroupRevokePermissionData{
		GroupUID:      group.UID,
		PermissionUID: permission.UID,
	})
}

// NewEventGroupPermissionsUpdated creates a new group permissions updated event.
func NewEventGroupPermissionsUpdated(group *model.Group) Event {
	return newEvent(EventGroupUpdatePermission, EventGroupUpdatePermissionData{
		GroupUID: group.UID,
	})
}
