package signal

import "time"

type SignalPermission struct {
	UID         *string
	Resource    *string
	Action      *string
	Description *string
	CreatedAt   *time.Time
	UpdatedAt   *time.Time

	// operation context
	Operation string
}
