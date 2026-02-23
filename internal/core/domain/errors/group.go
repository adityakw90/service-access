package errors

var (
	// Group not found error (23xxx range)
	ErrGroupNotFound = NewCustomError(23001, "group not found", nil)

	// Group operation errors (23xxx range)
	ErrGroupGetFailed              = NewCustomError(23002, "failed to get group", nil)
	ErrGroupUpdateFailed           = NewCustomError(23003, "failed to update group", nil)
	ErrGroupDeleteFailed           = NewCustomError(23004, "failed to delete group", nil)
	ErrGroupPermissionAssignFailed = NewCustomError(23005, "failed to assign permission to group", nil)
	ErrGroupPermissionRevokeFailed = NewCustomError(23006, "failed to revoke permission from group", nil)
	ErrGroupPermissionUpdateFailed = NewCustomError(23007, "failed to update group permissions", nil)
)
