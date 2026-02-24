package handler

import (
	"context"

	"github.com/adityakw90/service-access/internal/adapter/api/grpc/request"
	"github.com/adityakw90/service-access/internal/adapter/api/grpc/response"
	"github.com/adityakw90/service-access/internal/core/port/service"
	"github.com/adityakw90/service-access/internal/adapter/api/grpc/validator"
	"github.com/adityakw90/service-access-proto/gen/go/common"
	"github.com/adityakw90/service-access-proto/gen/go/group"
)

// GroupHandler must embed UnimplementedGroupServiceServer for forward compatibility
type GroupHandler struct {
	group.UnimplementedGroupServiceServer
	groupService service.GroupService
	validator    *validator.Validator
}

func NewGroupHandler(groupService service.GroupService, v *validator.Validator) *GroupHandler {
	return &GroupHandler{
		groupService: groupService,
		validator:    v,
	}
}

func (h *GroupHandler) Create(ctx context.Context, req *group.CreateRequest) (*group.CreateResponse, error) {
	param := request.ToGroupCreateParam(req)

	domainGroup, err := h.groupService.Create(ctx, param)
	if err != nil {
		return nil, err
	}

	return &group.CreateResponse{Uid: domainGroup.UID}, nil
}

func (h *GroupHandler) Get(ctx context.Context, req *group.GetRequest) (*group.Group, error) {
	domainGroup, err := h.groupService.Get(ctx, req.Uid)
	if err != nil {
		return nil, err
	}

	return response.ToProtoGroup(domainGroup), nil
}

func (h *GroupHandler) List(ctx context.Context, req *group.ListRequest) (*group.ListResponse, error) {
	paginationParam := request.ToPaginationParam(req.Pagination)
	filterParam := request.ToGroupListFilterParam(req)

	domainGroups, err := h.groupService.List(ctx, paginationParam, filterParam)
	if err != nil {
		return nil, err
	}

	return response.ToProtoGroupList(domainGroups, &domainGroups.Meta), nil
}

func (h *GroupHandler) Update(ctx context.Context, req *group.UpdateRequest) (*common.Success, error) {
	param := request.ToGroupUpdateParam(req)

	err := h.groupService.Update(ctx, req.Uid, param)
	if err != nil {
		return nil, err
	}

	return &common.Success{Success: true}, nil
}

func (h *GroupHandler) Delete(ctx context.Context, req *group.DeleteRequest) (*common.Success, error) {
	err := h.groupService.Delete(ctx, req.Uid)
	if err != nil {
		return nil, err
	}

	return &common.Success{Success: true}, nil
}

func (h *GroupHandler) ListPermissions(ctx context.Context, req *group.ListPermissionsRequest) (*group.ListPermissionsResponse, error) {
	paginationParam := request.ToPaginationParam(req.Pagination)
	filterParam := request.ToGroupPermissionListFilterParam(req)

	domainPerms, err := h.groupService.ListPermission(ctx, req.GroupUid, paginationParam, filterParam)
	if err != nil {
		return nil, err
	}

	return response.ToProtoGroupPermissionList(domainPerms, &domainPerms.Meta), nil
}

func (h *GroupHandler) UpdatePermission(ctx context.Context, req *group.UpdatePermissionRequest) (*common.Success, error) {
	err := h.groupService.UpdatePermission(ctx, req.GroupUid, req.PermissionUids)
	if err != nil {
		return nil, err
	}

	return &common.Success{Success: true}, nil
}

func (h *GroupHandler) AssignPermission(ctx context.Context, req *group.AssignPermissionRequest) (*common.Success, error) {
	err := h.groupService.AssignPermission(ctx, req.GroupUid, req.PermissionUid)
	if err != nil {
		return nil, err
	}

	return &common.Success{Success: true}, nil
}

func (h *GroupHandler) RevokePermission(ctx context.Context, req *group.RevokePermissionRequest) (*common.Success, error) {
	err := h.groupService.RevokePermission(ctx, req.GroupUid, req.PermissionUid)
	if err != nil {
		return nil, err
	}

	return &common.Success{Success: true}, nil
}
