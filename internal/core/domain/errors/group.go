package errors

var (
	ErrGroupNotFound = NewCustomError(20002, "group not found", nil)

	// Group operation errors
	ErrGroupGetFailed              = NewCustomError(20201, "failed to get group", nil)
	ErrGroupUpdateFailed           = NewCustomError(20202, "failed to update group", nil)
	ErrGroupDeleteFailed           = NewCustomError(20203, "failed to delete group", nil)
	ErrGroupPermissionAssignFailed = NewCustomError(20204, "failed to assign permission to group", nil)
	ErrGroupPermissionRevokeFailed = NewCustomError(20205, "failed to revoke permission from group", nil)
	ErrGroupPermissionUpdateFailed = NewCustomError(20206, "failed to update group permissions", nil)
)
