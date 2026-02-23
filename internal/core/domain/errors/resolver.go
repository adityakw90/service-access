package errors

// Resolver-related errors
var (
	ErrResolverUIDNotFound     = NewCustomError(20501, "resolver: UID not found in cache or database", nil)
	ErrResolverIDNotFound      = NewCustomError(20502, "resolver: ID not found in cache or database", nil)
	ErrResolverCacheFailure    = NewCustomError(20503, "resolver: cache operation failed", nil)
	ErrResolverDatabaseFailure = NewCustomError(20504, "resolver: database query failed", nil)
	ErrResolverInvalidInput    = NewCustomError(20505, "resolver: invalid input provided", nil)
)
