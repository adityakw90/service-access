package signal

import "time"

type SignalRole struct {
	UID         *string
	GroupUID    *string
	Name        *string
	Description *string
	CreatedAt   *time.Time
	UpdatedAt   *time.Time

	// operation context
	Operation string
}

type SignalRolePermission struct {
	RoleUID               *string
	GroupPermissionUID    *string
	PermissionUID         *string
	PermissionResource    *string
	PermissionAction      *string
	PermissionDescription *string
	CreatedAt             *time.Time

	// operation context
	Operation string
}
