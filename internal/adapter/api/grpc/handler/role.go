package handler

import (
	"context"

	"github.com/adityakw90/service-access/internal/adapter/api/grpc/request"
	"github.com/adityakw90/service-access/internal/adapter/api/grpc/response"
	"github.com/adityakw90/service-access/internal/core/port/service"
	"github.com/adityakw90/service-access/internal/adapter/api/grpc/validator"
	"github.com/adityakw90/service-access-proto/gen/go/common"
	"github.com/adityakw90/service-access-proto/gen/go/role"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RoleHandler must embed UnimplementedRoleServiceServer for forward compatibility
type RoleHandler struct {
	role.UnimplementedRoleServiceServer
	roleService service.RoleService
	validator   *validator.Validator
}

func NewRoleHandler(roleService service.RoleService, v *validator.Validator) *RoleHandler {
	return &RoleHandler{
		roleService: roleService,
		validator:   v,
	}
}

func (h *RoleHandler) Create(ctx context.Context, req *role.CreateRequest) (*role.CreateResponse, error) {
	// Convert and validate request using the request package
	roleReq := request.RoleCreateRequestFromPb(req)

	if err := h.validator.Struct(roleReq); err != nil {
		return nil, status.Error(codes.InvalidArgument, validator.ValidationErrors(err))
	}

	domainRole, err := h.roleService.Create(ctx, request.ToRoleCreateParam(req))
	if err != nil {
		return nil, err
	}

	return &role.CreateResponse{Uid: domainRole.UID}, nil
}

func (h *RoleHandler) Get(ctx context.Context, req *role.GetRequest) (*role.Role, error) {
	// Convert and validate request
	roleReq := request.RoleGetRequestFromPb(req)

	if err := h.validator.Struct(roleReq); err != nil {
		return nil, status.Error(codes.InvalidArgument, validator.ValidationErrors(err))
	}

	domainRole, err := h.roleService.Get(ctx, roleReq.UID)
	if err != nil {
		return nil, err
	}

	return response.ToProtoRole(domainRole), nil
}

func (h *RoleHandler) List(ctx context.Context, req *role.ListRequest) (*role.ListResponse, error) {
	paginationParam := request.ToPaginationParam(req.Pagination)
	filterParam := request.ToRoleListFilterParam(req)

	domainRoles, err := h.roleService.List(ctx, paginationParam, filterParam)
	if err != nil {
		return nil, err
	}

	return response.ToProtoRoleList(domainRoles, &domainRoles.Meta), nil
}

func (h *RoleHandler) Update(ctx context.Context, req *role.UpdateRequest) (*common.Success, error) {
	// Convert and validate request
	roleReq := request.RoleUpdateRequestFromPb(req)

	if err := h.validator.Struct(roleReq); err != nil {
		return nil, status.Error(codes.InvalidArgument, validator.ValidationErrors(err))
	}

	err := h.roleService.Update(ctx, roleReq.UID, request.ToRoleUpdateParam(req))
	if err != nil {
		return nil, err
	}

	return &common.Success{Success: true}, nil
}

func (h *RoleHandler) Delete(ctx context.Context, req *role.DeleteRequest) (*common.Success, error) {
	// Convert and validate request
	roleReq := request.RoleDeleteRequestFromPb(req)

	if err := h.validator.Struct(roleReq); err != nil {
		return nil, status.Error(codes.InvalidArgument, validator.ValidationErrors(err))
	}

	err := h.roleService.Delete(ctx, roleReq.UID)
	if err != nil {
		return nil, err
	}

	return &common.Success{Success: true}, nil
}

func (h *RoleHandler) ListPermissions(ctx context.Context, req *role.ListPermissionsRequest) (*role.ListPermissionsResponse, error) {
	paginationParam := request.ToPaginationParam(req.Pagination)
	filterParam := request.ToRolePermissionListFilterParam(req)

	domainPerms, err := h.roleService.ListPermission(ctx, req.RoleUid, paginationParam, filterParam)
	if err != nil {
		return nil, err
	}

	return response.ToProtoRolePermissionList(domainPerms, &domainPerms.Meta), nil
}

func (h *RoleHandler) UpdatePermission(ctx context.Context, req *role.UpdatePermissionRequest) (*common.Success, error) {
	// Convert and validate request
	roleReq := request.RoleUpdatePermissionRequestFromPb(req)

	if err := h.validator.Struct(roleReq); err != nil {
		return nil, status.Error(codes.InvalidArgument, validator.ValidationErrors(err))
	}

	err := h.roleService.UpdatePermission(ctx, roleReq.RoleUID, roleReq.GroupPermissionUIDs)
	if err != nil {
		return nil, err
	}

	return &common.Success{Success: true}, nil
}

func (h *RoleHandler) AssignPermission(ctx context.Context, req *role.AssignPermissionRequest) (*common.Success, error) {
	// Convert and validate request
	roleReq := request.RoleAssignPermissionRequestFromPb(req)

	if err := h.validator.Struct(roleReq); err != nil {
		return nil, status.Error(codes.InvalidArgument, validator.ValidationErrors(err))
	}

	err := h.roleService.AssignPermission(ctx, roleReq.RoleUID, roleReq.GroupPermissionUID)
	if err != nil {
		return nil, err
	}

	return &common.Success{Success: true}, nil
}

func (h *RoleHandler) RevokePermission(ctx context.Context, req *role.RevokePermissionRequest) (*common.Success, error) {
	// Convert and validate request
	roleReq := request.RoleRevokePermissionRequestFromPb(req)

	if err := h.validator.Struct(roleReq); err != nil {
		return nil, status.Error(codes.InvalidArgument, validator.ValidationErrors(err))
	}

	err := h.roleService.RevokePermission(ctx, roleReq.RoleUID, roleReq.GroupPermissionUID)
	if err != nil {
		return nil, err
	}

	return &common.Success{Success: true}, nil
}
