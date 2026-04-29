package errors

var (
	ErrAccessDenied = NewCustomError(7001, "access denied", nil)
)
