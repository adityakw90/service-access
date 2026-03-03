package handler

import (
	"context"

	"github.com/adityakw90/service-access/internal/adapter/api/grpc/request"
	"github.com/adityakw90/service-access/internal/adapter/api/grpc/response"
	"github.com/adityakw90/service-access/internal/core/port/service"
	"github.com/adityakw90/service-access/internal/adapter/api/grpc/validator"
	"github.com/adityakw90/service-access-proto/gen/go/common"
	"github.com/adityakw90/service-access-proto/gen/go/permission"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// PermissionHandler must embed UnimplementedPermissionServiceServer for forward compatibility
type PermissionHandler struct {
	permission.UnimplementedPermissionServiceServer
	permService service.PermissionService
	validator   *validator.Validator
}

func NewPermissionHandler(permService service.PermissionService, v *validator.Validator) *PermissionHandler {
	return &PermissionHandler{
		permService: permService,
		validator:   v,
	}
}

func (h *PermissionHandler) Create(ctx context.Context, req *permission.CreateRequest) (*permission.CreateResponse, error) {
	// Convert and validate request using the request package
	permReq := request.PermissionCreateRequestFromPb(req)

	if err := h.validator.Struct(permReq); err != nil {
		return nil, status.Error(codes.InvalidArgument, validator.ValidationErrors(err))
	}

	domainPerm, err := h.permService.Create(ctx, request.ToPermissionCreateParam(req))
	if err != nil {
		return nil, err
	}

	return &permission.CreateResponse{Uid: domainPerm.UID}, nil
}

func (h *PermissionHandler) Get(ctx context.Context, req *permission.GetRequest) (*permission.Permission, error) {
	// Convert and validate request
	permReq := request.PermissionGetRequestFromPb(req)

	if err := h.validator.Struct(permReq); err != nil {
		return nil, status.Error(codes.InvalidArgument, validator.ValidationErrors(err))
	}

	domainPerm, err := h.permService.Get(ctx, permReq.UID)
	if err != nil {
		return nil, err
	}

	return response.ToProtoPermission(domainPerm), nil
}

func (h *PermissionHandler) List(ctx context.Context, req *permission.ListRequest) (*permission.ListResponse, error) {
	paginationParam := request.ToPaginationParam(req.Pagination)
	filterParam := request.ToPermissionListFilterParam(req)

	domainPerms, err := h.permService.List(ctx, paginationParam, filterParam)
	if err != nil {
		return nil, err
	}

	return response.ToProtoPermissionList(domainPerms, &domainPerms.Meta), nil
}

func (h *PermissionHandler) Update(ctx context.Context, req *permission.UpdateRequest) (*common.Success, error) {
	// Convert and validate request
	permReq := request.PermissionUpdateRequestFromPb(req)

	if err := h.validator.Struct(permReq); err != nil {
		return nil, status.Error(codes.InvalidArgument, validator.ValidationErrors(err))
	}

	err := h.permService.Update(ctx, permReq.UID, request.ToPermissionUpdateParam(req))
	if err != nil {
		return nil, err
	}

	return &common.Success{Success: true}, nil
}

func (h *PermissionHandler) Delete(ctx context.Context, req *permission.DeleteRequest) (*common.Success, error) {
	// Convert and validate request
	permReq := request.PermissionDeleteRequestFromPb(req)

	if err := h.validator.Struct(permReq); err != nil {
		return nil, status.Error(codes.InvalidArgument, validator.ValidationErrors(err))
	}

	err := h.permService.Delete(ctx, permReq.UID)
	if err != nil {
		return nil, err
	}

	return &common.Success{Success: true}, nil
}
