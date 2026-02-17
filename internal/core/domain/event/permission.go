package event

import "time"

type EventPermissionCreateData struct {
	ID          int64
	UID         string
	Resource    string
	Action      string
	Description string
	CreatedAt   time.Time
}

type EventPermissionUpdateData struct {
	ID          int64
	UID         string
	Resource    string
	Action      string
	Description string
	UpdatedAt   time.Time
}

type EventPermissionDeleteData struct {
	ID  int64
	UID string
}
