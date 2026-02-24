package handler

import (
	"context"

	"github.com/adityakw90/service-access/internal/adapter/api/grpc/request"
	"github.com/adityakw90/service-access/internal/adapter/api/grpc/response"
	"github.com/adityakw90/service-access/internal/core/port/service"
	"github.com/adityakw90/service-access/internal/adapter/api/grpc/validator"
	"github.com/adityakw90/service-access-proto/gen/go/common"
	"github.com/adityakw90/service-access-proto/gen/go/role"
)

type RoleHandler struct {
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
	param := request.ToRoleCreateParam(req)

	domainRole, err := h.roleService.Create(ctx, param)
	if err != nil {
		return nil, err
	}

	return &role.CreateResponse{Uid: domainRole.UID}, nil
}

func (h *RoleHandler) Get(ctx context.Context, req *role.GetRequest) (*role.Role, error) {
	domainRole, err := h.roleService.Get(ctx, req.Uid)
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
	param := request.ToRoleUpdateParam(req)

	err := h.roleService.Update(ctx, req.Uid, param)
	if err != nil {
		return nil, err
	}

	return &common.Success{Success: true}, nil
}

func (h *RoleHandler) Delete(ctx context.Context, req *role.DeleteRequest) (*common.Success, error) {
	err := h.roleService.Delete(ctx, req.Uid)
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
	err := h.roleService.UpdatePermission(ctx, req.RoleUid, req.GroupPermissionUids)
	if err != nil {
		return nil, err
	}

	return &common.Success{Success: true}, nil
}

func (h *RoleHandler) AssignPermission(ctx context.Context, req *role.AssignPermissionRequest) (*common.Success, error) {
	err := h.roleService.AssignPermission(ctx, req.RoleUid, req.GroupPermissionUid)
	if err != nil {
		return nil, err
	}

	return &common.Success{Success: true}, nil
}

func (h *RoleHandler) RevokePermission(ctx context.Context, req *role.RevokePermissionRequest) (*common.Success, error) {
	err := h.roleService.RevokePermission(ctx, req.RoleUid, req.GroupPermissionUid)
	if err != nil {
		return nil, err
	}

	return &common.Success{Success: true}, nil
}
