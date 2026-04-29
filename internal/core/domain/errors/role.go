package errors

var (
	// Role not found error (21xxx range)
	ErrRoleNotFound = NewCustomError(21001, "role not found", nil)

	// Role operation errors (21xxx range)
	ErrRoleCreateFailed = NewCustomError(21002, "failed to create role", nil)
	ErrRoleGetFailed    = NewCustomError(21003, "failed to get role", nil)
	ErrRoleUpdateFailed = NewCustomError(21004, "failed to update role", nil)
	ErrRoleDeleteFailed = NewCustomError(21005, "failed to delete role", nil)
)
