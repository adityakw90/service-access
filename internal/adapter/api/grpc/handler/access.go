package handler

import (
	"context"

	"github.com/adityakw90/service-access/internal/core/port/service"
	"github.com/adityakw90/service-access-proto/gen/go/access"
)

// AccessHandler must embed UnimplementedAccessControlServiceServer for forward compatibility
type AccessHandler struct {
	access.UnimplementedAccessControlServiceServer
	accessService service.AccessService
}

func NewAccessHandler(accessSvc service.AccessService) *AccessHandler {
	return &AccessHandler{
		accessService: accessSvc,
	}
}

func (h *AccessHandler) CheckAccess(ctx context.Context, req *access.CheckAccessRequest) (*access.CheckAccessResponse, error) {
	allowed, reason, err := h.accessService.Check(ctx, req.SubjectId, req.SubjectType, req.Resource, req.Action)
	if err != nil {
		return nil, err
	}

	return &access.CheckAccessResponse{
		Allowed: allowed,
		Reason:  reason,
	}, nil
}
