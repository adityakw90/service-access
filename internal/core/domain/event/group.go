package event

import "time"

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
	GroupUID      string
	PermissionUID string
}
