package errors

var (
	ErrRoleNotFound = NewCustomError(20003, "role not found", nil)
)
