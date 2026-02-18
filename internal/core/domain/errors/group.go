package errors

var (
	ErrGroupNotFound = NewCustomError(20002, "group not found", nil)
)
