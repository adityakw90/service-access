package errors

var (
	ErrGroupPermissionNotFound = NewCustomError(20004, "group_permission not found", nil)
)
