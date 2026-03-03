package handler

import (
	"context"

	"github.com/adityakw90/service-access/internal/adapter/api/grpc/request"
	"github.com/adityakw90/service-access/internal/adapter/api/grpc/response"
	"github.com/adityakw90/service-access/internal/core/port/service"
	"github.com/adityakw90/service-access/internal/adapter/api/grpc/validator"
	"github.com/adityakw90/service-access-proto/gen/go/common"
	"github.com/adityakw90/service-access-proto/gen/go/subject"
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
	err := h.subjectService.Assign(ctx, req.SubjectId, req.SubjectType, req.RoleUid)
	if err != nil {
		return nil, err
	}

	return &common.Success{Success: true}, nil
}

func (h *SubjectHandler) RevokeRole(ctx context.Context, req *subject.RevokeRoleRequest) (*common.Success, error) {
	err := h.subjectService.Revoke(ctx, req.SubjectId, req.SubjectType, req.RoleUid)
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
