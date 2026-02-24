package response

import "github.com/adityakw90/service-access-proto/gen/go/access"

func ToAccessResponse(allowed bool, reason string) *access.CheckAccessResponse {
	return &access.CheckAccessResponse{
		Allowed: allowed,
		Reason:  reason,
	}
}
