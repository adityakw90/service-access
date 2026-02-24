package handler

import (
	"context"

	"github.com/adityakw90/service-access/internal/adapter/api/grpc/request"
	"github.com/adityakw90/service-access/internal/adapter/api/grpc/response"
	"github.com/adityakw90/service-access/internal/core/port/service"
	"github.com/adityakw90/service-access/internal/adapter/api/grpc/validator"
	"github.com/adityakw90/service-access-proto/gen/go/common"
	"github.com/adityakw90/service-access-proto/gen/go/permission"
)

type PermissionHandler struct {
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
	param := request.ToPermissionCreateParam(req)

	domainPerm, err := h.permService.Create(ctx, param)
	if err != nil {
		return nil, err
	}

	return &permission.CreateResponse{Uid: domainPerm.UID}, nil
}

func (h *PermissionHandler) Get(ctx context.Context, req *permission.GetRequest) (*permission.Permission, error) {
	domainPerm, err := h.permService.Get(ctx, req.Uid)
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
	param := request.ToPermissionUpdateParam(req)

	err := h.permService.Update(ctx, req.Uid, param)
	if err != nil {
		return nil, err
	}

	return &common.Success{Success: true}, nil
}

func (h *PermissionHandler) Delete(ctx context.Context, req *permission.DeleteRequest) (*common.Success, error) {
	err := h.permService.Delete(ctx, req.Uid)
	if err != nil {
		return nil, err
	}

	return &common.Success{Success: true}, nil
}
