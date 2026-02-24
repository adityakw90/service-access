package request

import (
	"strings"

	permission "github.com/adityakw90/service-access-proto/gen/go/permission"
	param "github.com/adityakw90/service-access/internal/core/domain/param"
)

// PermissionCreateRequest represents validated permission creation request data.
type PermissionCreateRequest struct {
	Resource    string `validate:"required"`
	Action      string `validate:"required"`
	Description string
}

// PermissionCreateRequestFromPb converts a proto CreateRequest to a PermissionCreateRequest.
func PermissionCreateRequestFromPb(req *permission.CreateRequest) *PermissionCreateRequest {
	return &PermissionCreateRequest{
		Resource:    strings.TrimSpace(req.GetResource()),
		Action:      strings.TrimSpace(req.GetAction()),
		Description: strings.TrimSpace(req.GetDescription()),
	}
}

// ToPermissionCreateParams converts a PermissionCreateRequest to domain params.
func (r *PermissionCreateRequest) ToPermissionCreateParams() *param.PermissionCreateParam {
	return &param.PermissionCreateParam{
		Resource:    r.Resource,
		Action:      r.Action,
		Description: r.Description,
	}
}

// PermissionUpdateRequest represents validated permission update request data.
type PermissionUpdateRequest struct {
	UID         string
	Resource    *string
	Action      *string
	Description *string
}

// PermissionUpdateRequestFromPb converts a proto UpdateRequest to a PermissionUpdateRequest.
func PermissionUpdateRequestFromPb(req *permission.UpdateRequest) *PermissionUpdateRequest {
	r := &PermissionUpdateRequest{
		UID: strings.TrimSpace(req.GetUid()),
	}

	if resource := strings.TrimSpace(req.GetResource()); resource != "" {
		r.Resource = &resource
	}
	if action := strings.TrimSpace(req.GetAction()); action != "" {
		r.Action = &action
	}
	if description := strings.TrimSpace(req.GetDescription()); description != "" {
		r.Description = &description
	}

	return r
}

// ToPermissionUpdateParams converts a PermissionUpdateRequest to domain params.
func (r *PermissionUpdateRequest) ToPermissionUpdateParams() *param.PermissionUpdateParam {
	return &param.PermissionUpdateParam{
		Resource:    r.Resource,
		Action:      r.Action,
		Description: r.Description,
	}
}

// GetUID returns the UID for update operations.
func (r *PermissionUpdateRequest) GetUID() string {
	return r.UID
}

// PermissionGetRequest represents validated permission get request data.
type PermissionGetRequest struct {
	UID string `validate:"required"`
}

// PermissionGetRequestFromPb converts a proto GetRequest to a PermissionGetRequest.
func PermissionGetRequestFromPb(req *permission.GetRequest) *PermissionGetRequest {
	return &PermissionGetRequest{
		UID: strings.TrimSpace(req.GetUid()),
	}
}

// GetUID returns the UID for get operations.
func (r *PermissionGetRequest) GetUID() string {
	return r.UID
}

// PermissionDeleteRequest represents validated permission delete request data.
type PermissionDeleteRequest struct {
	UID string `validate:"required"`
}

// PermissionDeleteRequestFromPb converts a proto DeleteRequest to a PermissionDeleteRequest.
func PermissionDeleteRequestFromPb(req *permission.DeleteRequest) *PermissionDeleteRequest {
	return &PermissionDeleteRequest{
		UID: strings.TrimSpace(req.GetUid()),
	}
}

// GetUID returns the UID for delete operations.
func (r *PermissionDeleteRequest) GetUID() string {
	return r.UID
}

// PermissionListRequest represents validated permission list request data.
type PermissionListRequest struct {
	Pagination *PaginationRequest
	Filter     *PermissionFilterRequest
}

// PermissionListRequestFromPb converts a proto ListRequest to a PermissionListRequest.
func PermissionListRequestFromPb(req *permission.ListRequest) *PermissionListRequest {
	payload := &PermissionListRequest{}

	if req.Pagination != nil {
		payload.Pagination = PaginationRequestFromPb(req.GetPagination())
	}

	if req.Filter != nil {
		payload.Filter = permissionFilterFromPb(req.GetFilter())
	}

	return payload
}

// ToPermissionListParams converts a PermissionListRequest to domain params.
func (r *PermissionListRequest) ToPermissionListParams() *param.PermissionListParam {
	var pagination *param.PaginationParam
	if r.Pagination != nil {
		pagination = r.Pagination.ToPaginationParam()
	}

	var filter *param.PermissionListFilterParam
	if r.Filter != nil {
		filter = r.Filter.toPermissionListFilterParam()
	}

	return &param.PermissionListParam{
		Pagination: pagination,
		Filter:     filter,
	}
}

// PermissionFilterRequest represents validated permission filter request data.
type PermissionFilterRequest struct {
	UIDs     []string
	Resource *string
	Action   *string
	Query    *string
}

// permissionFilterFromPb converts a proto FilterRequest to a PermissionFilterRequest.
func permissionFilterFromPb(req *permission.FilterRequest) *PermissionFilterRequest {
	if req == nil {
		return nil
	}

	r := &PermissionFilterRequest{
		UIDs: req.GetUids(),
	}

	if resource := strings.TrimSpace(req.GetResource()); resource != "" {
		r.Resource = &resource
	}
	if action := strings.TrimSpace(req.GetAction()); action != "" {
		r.Action = &action
	}
	if query := strings.TrimSpace(req.GetQuery()); query != "" {
		r.Query = &query
	}

	return r
}

// toPermissionListFilterParam converts a PermissionFilterRequest to domain params.
func (r *PermissionFilterRequest) toPermissionListFilterParam() *param.PermissionListFilterParam {
	return &param.PermissionListFilterParam{
		UIDs:     r.UIDs,
		Resource: r.Resource,
		Action:   r.Action,
		Query:    r.Query,
	}
}
