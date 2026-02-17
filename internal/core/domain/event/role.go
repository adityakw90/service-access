package event

import "time"

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
	RoleUID            string
	GroupPermissionUID string
}
