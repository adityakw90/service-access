package event

import "time"

type EventGroupCreateData struct {
	ID          int64
	UID         string
	Name        string
	Description string
	CreatedAt   time.Time
}

type EventGroupUpdateData struct {
	ID          int64
	UID         string
	Name        string
	Description string
	UpdatedAt   time.Time
}

type EventGroupDeleteData struct {
	ID  int64
	UID string
}

type EventGroupAssignPermissionData struct {
	GroupID       int64
	GroupUID      string
	PermissionID  int64
	PermissionUID string
}

type EventGroupRevokePermissionData struct {
	GroupID       int64
	GroupUID      string
	PermissionID  int64
	PermissionUID string
}

type EventGroupUpdatePermissionData struct {
	GroupID       int64
	GroupUID      string
	PermissionID  int64
	PermissionUID string
}
