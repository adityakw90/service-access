package response

import (
	domainErrors "github.com/adityakw90/service-access/internal/core/domain/errors"
	"google.golang.org/grpc/codes"
)

var GrpcErrorCode = map[int]codes.Code{
	domainErrors.ErrInternalServerError.Code:     codes.Internal,
	domainErrors.ErrTraceInformationMissing.Code: codes.Internal,
	domainErrors.ErrRequestCanceled.Code:         codes.Canceled,
	domainErrors.ErrRequestTimeout.Code:          codes.DeadlineExceeded,
	domainErrors.ErrRequestAborted.Code:          codes.Aborted,
	domainErrors.ErrUnimplemented.Code:           codes.Unimplemented,
	domainErrors.ErrNotFound.Code:                codes.NotFound,
	domainErrors.ErrInvalidArgument.Code:         codes.InvalidArgument,
	domainErrors.ErrValidation.Code:              codes.InvalidArgument,
	domainErrors.ErrPermissionDenied.Code:        codes.PermissionDenied,
	domainErrors.ErrResourceConflict.Code:        codes.Aborted,
}

// Extract the status code (first 3 digits)
func extractErrorCode(code int) codes.Code {
	if grpcCode, ok := GrpcErrorCode[code]; ok {
		return grpcCode
	}
	return codes.Internal
}
