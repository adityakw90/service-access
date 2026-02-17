package model

import "time"

// Permission represents a permission entity in the domain layer.
type Permission struct {
	ID          int64
	UID         string
	Resource    string
	Action      string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Permissions contains the list of permissions and metadata for pagination.
type Permissions struct {
	Items []Permission
	Meta  Meta
}
