package handler

import (
	"context"

	"github.com/adityakw90/service-access/internal/adapter/api/grpc/request"
	"github.com/adityakw90/service-access/internal/adapter/api/grpc/response"
	"github.com/adityakw90/service-access/internal/core/port/service"
	"github.com/adityakw90/service-access/internal/adapter/api/grpc/validator"
	"github.com/adityakw90/service-access-proto/gen/go/common"
	"github.com/adityakw90/service-access-proto/gen/go/group"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	// Convert and validate request using the request package
	groupReq := request.GroupCreateRequestFromPb(req)

	if err := h.validator.Struct(groupReq); err != nil {
		return nil, status.Error(codes.InvalidArgument, validator.ValidationErrors(err))
	}

	domainGroup, err := h.groupService.Create(ctx, request.ToGroupCreateParam(req))
	if err != nil {
		return nil, err
	}

	return &group.CreateResponse{Uid: domainGroup.UID}, nil
}

func (h *GroupHandler) Get(ctx context.Context, req *group.GetRequest) (*group.Group, error) {
	// Convert and validate request
	groupReq := request.GroupGetRequestFromPb(req)

	if err := h.validator.Struct(groupReq); err != nil {
		return nil, status.Error(codes.InvalidArgument, validator.ValidationErrors(err))
	}

	domainGroup, err := h.groupService.Get(ctx, groupReq.UID)
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
	// Convert and validate request
	groupReq := request.GroupUpdateRequestFromPb(req)

	if err := h.validator.Struct(groupReq); err != nil {
		return nil, status.Error(codes.InvalidArgument, validator.ValidationErrors(err))
	}

	err := h.groupService.Update(ctx, groupReq.UID, request.ToGroupUpdateParam(req))
	if err != nil {
		return nil, err
	}

	return &common.Success{Success: true}, nil
}

func (h *GroupHandler) Delete(ctx context.Context, req *group.DeleteRequest) (*common.Success, error) {
	// Convert and validate request
	groupReq := request.GroupDeleteRequestFromPb(req)

	if err := h.validator.Struct(groupReq); err != nil {
		return nil, status.Error(codes.InvalidArgument, validator.ValidationErrors(err))
	}

	err := h.groupService.Delete(ctx, groupReq.UID)
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
	// Convert and validate request
	groupReq := request.GroupUpdatePermissionRequestFromPb(req)

	if err := h.validator.Struct(groupReq); err != nil {
		return nil, status.Error(codes.InvalidArgument, validator.ValidationErrors(err))
	}

	err := h.groupService.UpdatePermission(ctx, groupReq.GroupUID, groupReq.PermissionUIDs)
	if err != nil {
		return nil, err
	}

	return &common.Success{Success: true}, nil
}

func (h *GroupHandler) AssignPermission(ctx context.Context, req *group.AssignPermissionRequest) (*common.Success, error) {
	// Convert and validate request
	groupReq := request.GroupAssignPermissionRequestFromPb(req)

	if err := h.validator.Struct(groupReq); err != nil {
		return nil, status.Error(codes.InvalidArgument, validator.ValidationErrors(err))
	}

	err := h.groupService.AssignPermission(ctx, groupReq.GroupUID, groupReq.PermissionUID)
	if err != nil {
		return nil, err
	}

	return &common.Success{Success: true}, nil
}

func (h *GroupHandler) RevokePermission(ctx context.Context, req *group.RevokePermissionRequest) (*common.Success, error) {
	// Convert and validate request
	groupReq := request.GroupRevokePermissionRequestFromPb(req)

	if err := h.validator.Struct(groupReq); err != nil {
		return nil, status.Error(codes.InvalidArgument, validator.ValidationErrors(err))
	}

	err := h.groupService.RevokePermission(ctx, groupReq.GroupUID, groupReq.PermissionUID)
	if err != nil {
		return nil, err
	}

	return &common.Success{Success: true}, nil
}
