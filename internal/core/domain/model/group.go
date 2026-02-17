package model

import "time"

// Group represents a group of users and permissions.
type Group struct {
	ID          int64
	UID         string
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Groups contains the list of groups and metadata for pagination.
type Groups struct {
	Items []Group
	Meta  Meta
}

// GroupPermission represents the association between a group and a permission.
type GroupPermission struct {
	ID                    int64
	UID                   string
	GroupID               int64
	GroupUID              string
	PermissionID          int64
	PermissionUID         string
	PermissionResource    string
	PermissionAction      string
	PermissionDescription string
	CreatedAt             time.Time
}
