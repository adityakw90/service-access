package errors

var (
	ErrRoleNotFound = NewCustomError(20003, "role not found", nil)

	// Role operation errors
	ErrRoleGetFailed    = NewCustomError(20401, "failed to get role", nil)
	ErrRoleUpdateFailed = NewCustomError(20402, "failed to update role", nil)
	ErrRoleDeleteFailed = NewCustomError(20403, "failed to delete role", nil)
)
