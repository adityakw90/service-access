package event

import "time"

type EventRoleCreateData struct {
	ID          int64
	UID         string
	GroupID     int64
	GroupUID    string
	Name        string
	Description string
	CreatedAt   time.Time
}

type EventRoleUpdateData struct {
	ID          int64
	UID         string
	Name        string
	Description string
	UpdatedAt   time.Time
}

type EventRoleDeleteData struct {
	ID  int64
	UID string
}

type EventRoleAssignPermissionData struct {
	RoleID             int64
	RoleUID            string
	GroupPermissionID  int64
	GroupPermissionUID string
}

type EventRoleRevokePermissionData struct {
	RoleID             int64
	RoleUID            string
	GroupPermissionID  int64
	GroupPermissionUID string
}

type EventRoleUpdatePermissionData struct {
	RoleID             int64
	RoleUID            string
	GroupPermissionID  int64
	GroupPermissionUID string
}
