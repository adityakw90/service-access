package errors

// Resolver-related errors (20xxx range)
var (
	ErrResolverUIDNotFound     = NewCustomError(20001, "resolver: UID not found in cache or database", nil)
	ErrResolverIDNotFound      = NewCustomError(20002, "resolver: ID not found in cache or database", nil)
	ErrResolverCacheFailure    = NewCustomError(20003, "resolver: cache operation failed", nil)
	ErrResolverDatabaseFailure = NewCustomError(20004, "resolver: database query failed", nil)
	ErrResolverInvalidInput    = NewCustomError(20005, "resolver: invalid input provided", nil)
)
