package errors

// GroupPermission-related errors (23xxx range)
var (
	ErrGroupPermissionNotFound = NewCustomError(23008, "group_permission not found", nil)
)
