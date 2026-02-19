package errors

var (
	ErrPermissionNotFound = NewCustomError(20001, "permission not found", nil)
)
