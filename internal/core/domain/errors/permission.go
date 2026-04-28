package errors

var (
	// Permission not found error (22xxx range)
	ErrPermissionNotFound = NewCustomError(22001, "permission not found", nil)

	// Permission operation errors (22xxx range)
	ErrPermissionGetFailed    = NewCustomError(22002, "failed to get permission", nil)
	ErrPermissionUpdateFailed = NewCustomError(22003, "failed to update permission", nil)
	ErrPermissionDeleteFailed = NewCustomError(22004, "failed to delete permission", nil)
)
