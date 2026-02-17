package event

import "time"

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
