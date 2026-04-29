package model

import "time"

// Role represents a role within a group.
type Role struct {
	ID          int64
	UID         string
	GroupID     int64
	GroupUID    string
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Roles contains the list of roles and metadata for pagination.
type Roles struct {
	Items []Role
	Meta  Meta
}

// RolePermission represents the association between a role and a group permission.
type RolePermission struct {
	RoleID                int64
	RoleUID               string
	GroupPermissionID     int64
	GroupPermissionUID    string
	PermissionUID         string
	PermissionResource    string
	PermissionAction      string
	PermissionDescription string
	CreatedAt             time.Time
}

// RolePermissions contains the list of role permissions and metadata for pagination.
type RolePermissions struct {
	Items []RolePermission
	Meta  Meta
}
