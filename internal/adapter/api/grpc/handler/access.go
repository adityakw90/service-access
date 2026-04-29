package handler

import (
	"context"

	"github.com/adityakw90/service-access-proto/gen/go/access"
	"github.com/adityakw90/service-access/internal/adapter/api/grpc/request"
	"github.com/adityakw90/service-access/internal/adapter/api/grpc/validator"
	"github.com/adityakw90/service-access/internal/core/port/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AccessHandler must embed UnimplementedAccessControlServiceServer for forward compatibility
type AccessHandler struct {
	access.UnimplementedAccessControlServiceServer
	accessService service.AccessService
	validator     *validator.Validator
}

func NewAccessHandler(accessSvc service.AccessService, v *validator.Validator) *AccessHandler {
	return &AccessHandler{
		accessService: accessSvc,
		validator:     v,
	}
}

func (h *AccessHandler) CheckAccess(ctx context.Context, req *access.CheckAccessRequest) (*access.CheckAccessResponse, error) {
	// Convert and validate request using the request package
	accessReq := request.AccessRequestFromPb(req)

	if err := h.validator.Struct(accessReq); err != nil {
		return nil, status.Error(codes.InvalidArgument, validator.ValidationErrors(err))
	}

	allowed, reason, err := h.accessService.Check(ctx, accessReq.SubjectID, accessReq.SubjectType, accessReq.Resource, accessReq.Action)
	if err != nil {
		return nil, err
	}

	return &access.CheckAccessResponse{
		Allowed: allowed,
		Reason:  reason,
	}, nil
}
