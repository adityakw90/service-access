package signal

import "time"

type SignalGroup struct {
	UID         *string
	Name        *string
	Description *string
	CreatedAt   *time.Time
	UpdatedAt   *time.Time

	// operation context
	Operation string
}

type SignalGroupPermission struct {
	UID                   *string
	GroupUID              *string
	PermissionUID         *string
	PermissionResource    *string
	PermissionAction      *string
	PermissionDescription *string
	CreatedAt             *time.Time

	// operation context
	Operation string
}
