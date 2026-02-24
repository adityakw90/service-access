package handler

import (
	"context"

	"github.com/adityakw90/service-access/internal/adapter/api/grpc/request"
	"github.com/adityakw90/service-access/internal/adapter/api/grpc/response"
	"github.com/adityakw90/service-access/internal/core/port/service"
	"github.com/adityakw90/service-access/internal/adapter/api/grpc/validator"
	"github.com/adityakw90/service-access-proto/gen/go/common"
	"github.com/adityakw90/service-access-proto/gen/go/access"
)

type AccessHandler struct {
	accessService  service.AccessService
	subjectService service.SubjectService
	validator      *validator.Validator
}

func NewAccessHandler(accessSvc service.AccessService, subjectSvc service.SubjectService, v *validator.Validator) *AccessHandler {
	return &AccessHandler{
		accessService:  accessSvc,
		subjectService: subjectSvc,
		validator:      v,
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

func (h *AccessHandler) AssignRole(ctx context.Context, req *access.AssignRoleRequest) (*common.Success, error) {
	err := h.subjectService.Assign(ctx, req.SubjectId, req.SubjectType, req.RoleUid)
	if err != nil {
		return nil, err
	}

	return &common.Success{Success: true}, nil
}

func (h *AccessHandler) RevokeRole(ctx context.Context, req *access.RevokeRoleRequest) (*common.Success, error) {
	err := h.subjectService.Revoke(ctx, req.SubjectId, req.SubjectType, req.RoleUid)
	if err != nil {
		return nil, err
	}

	return &common.Success{Success: true}, nil
}

func (h *AccessHandler) ListSubjectRoles(ctx context.Context, req *access.ListSubjectRolesRequest) (*access.ListSubjectRolesResponse, error) {
	paginationParam := request.ToPaginationParam(req.Pagination)
	filterParam := request.ToAccessListFilterParam(req)

	subjectRoles, err := h.subjectService.List(ctx, paginationParam, filterParam)
	if err != nil {
		return nil, err
	}

	return response.ToProtoSubjectRoleList(subjectRoles, &subjectRoles.Meta), nil
}
