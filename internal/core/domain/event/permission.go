package event

import (
	"time"

	"github.com/adityakw90/service-access/internal/core/domain/model"
)

type EventPermissionCreateData struct {
	UID         string
	Resource    string
	Action      string
	Description string
	CreatedAt   time.Time
}

type EventPermissionUpdateData struct {
	UID         string
	Resource    string
	Action      string
	Description string
	UpdatedAt   time.Time
}

type EventPermissionDeleteData struct {
	UID string
}

// NewEventPermissionCreated creates a new permission created event.
func NewEventPermissionCreated(permission *model.Permission) Event {
	return newEvent(EventPermissionCreate, EventPermissionCreateData{
		UID:         permission.UID,
		Resource:    permission.Resource,
		Action:      permission.Action,
		Description: permission.Description,
		CreatedAt:   permission.CreatedAt,
	})
}

// NewEventPermissionUpdated creates a new permission updated event.
func NewEventPermissionUpdated(permission *model.Permission) Event {
	return newEvent(EventPermissionUpdate, EventPermissionUpdateData{
		UID:         permission.UID,
		Resource:    permission.Resource,
		Action:      permission.Action,
		Description: permission.Description,
		UpdatedAt:   permission.UpdatedAt,
	})
}

// NewEventPermissionDeleted creates a new permission deleted event.
func NewEventPermissionDeleted(permission *model.Permission) Event {
	return newEvent(EventPermissionDelete, EventPermissionDeleteData{
		UID: permission.UID,
	})
}
