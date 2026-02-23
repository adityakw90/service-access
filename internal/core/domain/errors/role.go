package errors

var (
	// Role not found error (21xxx range)
	ErrRoleNotFound = NewCustomError(21001, "role not found", nil)

	// Role operation errors (21xxx range)
	ErrRoleGetFailed    = NewCustomError(21002, "failed to get role", nil)
	ErrRoleUpdateFailed = NewCustomError(21003, "failed to update role", nil)
	ErrRoleDeleteFailed = NewCustomError(21004, "failed to delete role", nil)
)
