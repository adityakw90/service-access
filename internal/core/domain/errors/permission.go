package errors

var (
	ErrPermissionNotFound = NewCustomError(20001, "permission not found", nil)

	// Permission operation errors
	ErrPermissionGetFailed    = NewCustomError(20301, "failed to get permission", nil)
	ErrPermissionUpdateFailed = NewCustomError(20302, "failed to update permission", nil)
	ErrPermissionDeleteFailed = NewCustomError(20303, "failed to delete permission", nil)
)
