package handler

import (
	"context"

	"github.com/adityakw90/service-access-proto/gen/go/common"
	"github.com/adityakw90/service-access-proto/gen/go/group"
	"github.com/adityakw90/service-access-proto/gen/go/permission"
	"github.com/adityakw90/service-access-proto/gen/go/role"
	"github.com/adityakw90/service-access-proto/gen/go/subject"
	"github.com/adityakw90/service-access/internal/adapter/api/grpc/request"
	"github.com/adityakw90/service-access/internal/adapter/api/grpc/response"
	"github.com/adityakw90/service-access/internal/adapter/api/grpc/validator"
	"github.com/adityakw90/service-access/internal/core/port/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SubjectHandler must embed UnimplementedSubjectServiceServer for forward compatibility
type SubjectHandler struct {
	subject.UnimplementedSubjectServiceServer
	subjectService service.SubjectService
	validator      *validator.Validator
}

func NewSubjectHandler(subjectSvc service.SubjectService, v *validator.Validator) *SubjectHandler {
	return &SubjectHandler{
		subjectService: subjectSvc,
		validator:      v,
	}
}

func (h *SubjectHandler) AssignRole(ctx context.Context, req *subject.AssignRoleRequest) (*common.Success, error) {
	// Convert and validate request using the request package
	subjectReq := request.AssignRoleRequestFromPb(req)

	if err := h.validator.Struct(subjectReq); err != nil {
		return nil, status.Error(codes.InvalidArgument, validator.ValidationErrors(err))
	}

	err := h.subjectService.Assign(ctx, subjectReq.SubjectID, subjectReq.SubjectType, subjectReq.RoleUID)
	if err != nil {
		return nil, err
	}

	return &common.Success{Success: true}, nil
}

func (h *SubjectHandler) RevokeRole(ctx context.Context, req *subject.RevokeRoleRequest) (*common.Success, error) {
	// Convert and validate request using the request package
	subjectReq := request.RevokeRoleRequestFromPb(req)

	if err := h.validator.Struct(subjectReq); err != nil {
		return nil, status.Error(codes.InvalidArgument, validator.ValidationErrors(err))
	}

	err := h.subjectService.Revoke(ctx, subjectReq.SubjectID, subjectReq.SubjectType, subjectReq.RoleUID)
	if err != nil {
		return nil, err
	}

	return &common.Success{Success: true}, nil
}

func (h *SubjectHandler) List(ctx context.Context, req *subject.ListRequest) (*subject.ListResponse, error) {
	paginationParam := request.ToPaginationParam(req.Pagination)
	filterParam := request.ToSubjectListFilterParam(req)

	subjectRoles, err := h.subjectService.List(ctx, paginationParam, filterParam)
	if err != nil {
		return nil, err
	}

	return response.ToProtoSubjectRoleList(subjectRoles, &subjectRoles.Meta), nil
}

// Get returns full subject profile with groups, roles, and permissions
func (h *SubjectHandler) Get(ctx context.Context, req *subject.GetSubjectRequest) (*subject.GetSubjectResponse, error) {
	subjectReq := request.GetSubjectRequestFromPb(req)
	if err := h.validator.Struct(subjectReq); err != nil {
		return nil, status.Error(codes.InvalidArgument, validator.ValidationErrors(err))
	}

	profile, err := h.subjectService.GetFullProfile(ctx, subjectReq.SubjectID, subjectReq.SubjectType)
	if err != nil {
		return nil, err
	}

	resp := &subject.GetSubjectResponse{
		Groups:          make([]*group.Group, 0, len(profile.Groups)),
		Roles:           make([]*role.Role, 0, len(profile.Roles)),
		Permissions:     make([]*permission.Permission, 0, len(profile.Permissions)),
		TotalGroup:      int32(len(profile.Groups)),
		TotalRole:       int32(len(profile.Roles)),
		TotalPermission: int32(len(profile.Permissions)),
	}

	for _, g := range profile.Groups {
		resp.Groups = append(resp.Groups, response.ToProtoGroup(&g))
	}
	for _, r := range profile.Roles {
		resp.Roles = append(resp.Roles, response.ToProtoRole(&r))
	}
	for _, p := range profile.Permissions {
		resp.Permissions = append(resp.Permissions, response.ToProtoPermission(&p))
	}

	return resp, nil
}

// ListGroup returns all groups for a subject
func (h *SubjectHandler) ListGroup(ctx context.Context, req *subject.GetSubjectRequest) (*subject.ListGroupResponse, error) {
	subjectReq := request.GetSubjectRequestFromPb(req)
	if err := h.validator.Struct(subjectReq); err != nil {
		return nil, status.Error(codes.InvalidArgument, validator.ValidationErrors(err))
	}

	groups, err := h.subjectService.GetGroups(ctx, subjectReq.SubjectID, subjectReq.SubjectType)
	if err != nil {
		return nil, err
	}

	resp := &subject.ListGroupResponse{
		Groups: make([]*group.Group, 0, len(groups)),
		Total:  int32(len(groups)),
	}

	for _, g := range groups {
		resp.Groups = append(resp.Groups, response.ToProtoGroup(&g))
	}

	return resp, nil
}

// ListRole returns all roles for a subject (direct and via groups)
func (h *SubjectHandler) ListRole(ctx context.Context, req *subject.GetSubjectRequest) (*subject.ListRoleResponse, error) {
	subjectReq := request.GetSubjectRequestFromPb(req)
	if err := h.validator.Struct(subjectReq); err != nil {
		return nil, status.Error(codes.InvalidArgument, validator.ValidationErrors(err))
	}

	roles, err := h.subjectService.GetRoles(ctx, subjectReq.SubjectID, subjectReq.SubjectType)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &subject.ListRoleResponse{
		Roles: make([]*role.Role, 0, len(roles)),
		Total: int32(len(roles)),
	}

	for _, r := range roles {
		resp.Roles = append(resp.Roles, response.ToProtoRole(&r))
	}

	return resp, nil
}

// ListPermission returns all unique permissions for a subject
func (h *SubjectHandler) ListPermission(ctx context.Context, req *subject.GetSubjectRequest) (*subject.ListPermissionResponse, error) {
	subjectReq := request.GetSubjectRequestFromPb(req)
	if err := h.validator.Struct(subjectReq); err != nil {
		return nil, status.Error(codes.InvalidArgument, validator.ValidationErrors(err))
	}

	perms, err := h.subjectService.GetPermissions(ctx, subjectReq.SubjectID, subjectReq.SubjectType)
	if err != nil {
		return nil, err
	}

	resp := &subject.ListPermissionResponse{
		Permissions: make([]*permission.Permission, 0, len(perms)),
		Total:       int32(len(perms)),
	}

	for _, p := range perms {
		resp.Permissions = append(resp.Permissions, response.ToProtoPermission(&p))
	}

	return resp, nil
}
