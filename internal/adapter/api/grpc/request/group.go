package request

import (
	"strings"

	group "github.com/adityakw90/service-access-proto/gen/go/group"
	param "github.com/adityakw90/service-access/internal/core/domain/param"
)

// GroupCreateRequest represents validated group creation request data.
type GroupCreateRequest struct {
	Name        string `validate:"required"`
	Description string
}

// GroupCreateRequestFromPb converts a proto CreateRequest to a GroupCreateRequest.
func GroupCreateRequestFromPb(req *group.CreateRequest) *GroupCreateRequest {
	return &GroupCreateRequest{
		Name:        strings.TrimSpace(req.GetName()),
		Description: strings.TrimSpace(req.GetDescription()),
	}
}

// ToGroupCreateParams converts a GroupCreateRequest to domain params.
func (r *GroupCreateRequest) ToGroupCreateParams() *param.GroupCreateParam {
	return &param.GroupCreateParam{
		Name:        r.Name,
		Description: r.Description,
	}
}

// GroupUpdateRequest represents validated group update request data.
type GroupUpdateRequest struct {
	UID         string  `validate:"required"`
	Name        *string `validate:"omitempty"`
	Description *string `validate:"omitempty"`
}

// GroupUpdateRequestFromPb converts a proto UpdateRequest to a GroupUpdateRequest.
func GroupUpdateRequestFromPb(req *group.UpdateRequest) *GroupUpdateRequest {
	r := &GroupUpdateRequest{
		UID: strings.TrimSpace(req.GetUid()),
	}

	if name := strings.TrimSpace(req.GetName()); name != "" {
		r.Name = &name
	}
	if description := strings.TrimSpace(req.GetDescription()); description != "" {
		r.Description = &description
	}

	return r
}

// ToGroupUpdateParams converts a GroupUpdateRequest to domain params.
func (r *GroupUpdateRequest) ToGroupUpdateParams() *param.GroupUpdateParam {
	return &param.GroupUpdateParam{
		Name:        r.Name,
		Description: r.Description,
	}
}

// GetUID returns the UID for update operations.
func (r *GroupUpdateRequest) GetUID() string {
	return r.UID
}

// GroupGetRequest represents validated group get request data.
type GroupGetRequest struct {
	UID string `validate:"required"`
}

// GroupGetRequestFromPb converts a proto GetRequest to a GroupGetRequest.
func GroupGetRequestFromPb(req *group.GetRequest) *GroupGetRequest {
	return &GroupGetRequest{
		UID: strings.TrimSpace(req.GetUid()),
	}
}

// GetUID returns the UID for get operations.
func (r *GroupGetRequest) GetUID() string {
	return r.UID
}

// GroupDeleteRequest represents validated group delete request data.
type GroupDeleteRequest struct {
	UID string `validate:"required"`
}

// GroupDeleteRequestFromPb converts a proto DeleteRequest to a GroupDeleteRequest.
func GroupDeleteRequestFromPb(req *group.DeleteRequest) *GroupDeleteRequest {
	return &GroupDeleteRequest{
		UID: strings.TrimSpace(req.GetUid()),
	}
}

// GetUID returns the UID for delete operations.
func (r *GroupDeleteRequest) GetUID() string {
	return r.UID
}

// GroupListRequest represents validated group list request data.
type GroupListRequest struct {
	Pagination *PaginationRequest
	Filter     *GroupFilterRequest
}

// GroupListRequestFromPb converts a proto ListRequest to a GroupListRequest.
func GroupListRequestFromPb(req *group.ListRequest) *GroupListRequest {
	payload := &GroupListRequest{}

	if req.Pagination != nil {
		payload.Pagination = PaginationRequestFromPb(req.GetPagination())
	}

	if req.Filter != nil {
		payload.Filter = groupFilterFromPb(req.GetFilter())
	}

	return payload
}

// ToGroupListParams converts a GroupListRequest to domain params.
func (r *GroupListRequest) ToGroupListParams() *param.GroupListParam {
	var pagination *param.PaginationParam
	if r.Pagination != nil {
		pagination = r.Pagination.ToPaginationParam()
	}

	var filter *param.GroupListFilterParam
	if r.Filter != nil {
		filter = r.Filter.toGroupListFilterParam()
	}

	return &param.GroupListParam{
		Pagination: pagination,
		Filter:     filter,
	}
}

// GroupFilterRequest represents validated group filter request data.
type GroupFilterRequest struct {
	UIDs  []string
	Name  *string
	Query *string
}

// groupFilterFromPb converts a proto FilterRequest to a GroupFilterRequest.
func groupFilterFromPb(req *group.FilterRequest) *GroupFilterRequest {
	if req == nil {
		return nil
	}

	r := &GroupFilterRequest{
		UIDs: req.GetUids(),
	}

	if name := strings.TrimSpace(req.GetName()); name != "" {
		r.Name = &name
	}
	if query := strings.TrimSpace(req.GetQuery()); query != "" {
		r.Query = &query
	}

	return r
}

// toGroupListFilterParam converts a GroupFilterRequest to domain params.
func (r *GroupFilterRequest) toGroupListFilterParam() *param.GroupListFilterParam {
	return &param.GroupListFilterParam{
		UIDs:  r.UIDs,
		Name:  r.Name,
		Query: r.Query,
	}
}

// GroupUpdatePermissionRequest represents validated group permission update request data.
type GroupUpdatePermissionRequest struct {
	GroupUID       string `validate:"required"`
	PermissionUIDs []string `validate:"required,min=1"`
}

// GroupUpdatePermissionRequestFromPb converts a proto UpdatePermissionRequest to a GroupUpdatePermissionRequest.
func GroupUpdatePermissionRequestFromPb(req *group.UpdatePermissionRequest) *GroupUpdatePermissionRequest {
	return &GroupUpdatePermissionRequest{
		GroupUID:       strings.TrimSpace(req.GetGroupUid()),
		PermissionUIDs: req.GetPermissionUids(),
	}
}

// GetGroupUID returns the group UID.
func (r *GroupUpdatePermissionRequest) GetGroupUID() string {
	return r.GroupUID
}

// GetPermissionUIDs returns the permission UIDs.
func (r *GroupUpdatePermissionRequest) GetPermissionUIDs() []string {
	return r.PermissionUIDs
}

// GroupAssignPermissionRequest represents validated group permission assignment request data.
type GroupAssignPermissionRequest struct {
	GroupUID      string `validate:"required"`
	PermissionUID string `validate:"required"`
}

// GroupAssignPermissionRequestFromPb converts a proto AssignPermissionRequest to a GroupAssignPermissionRequest.
func GroupAssignPermissionRequestFromPb(req *group.AssignPermissionRequest) *GroupAssignPermissionRequest {
	return &GroupAssignPermissionRequest{
		GroupUID:      strings.TrimSpace(req.GetGroupUid()),
		PermissionUID: strings.TrimSpace(req.GetPermissionUid()),
	}
}

// GetGroupUID returns the group UID.
func (r *GroupAssignPermissionRequest) GetGroupUID() string {
	return r.GroupUID
}

// GetPermissionUID returns the permission UID.
func (r *GroupAssignPermissionRequest) GetPermissionUID() string {
	return r.PermissionUID
}

// GroupRevokePermissionRequest represents validated group permission revocation request data.
type GroupRevokePermissionRequest struct {
	GroupUID      string `validate:"required"`
	PermissionUID string `validate:"required"`
}

// GroupRevokePermissionRequestFromPb converts a proto RevokePermissionRequest to a GroupRevokePermissionRequest.
func GroupRevokePermissionRequestFromPb(req *group.RevokePermissionRequest) *GroupRevokePermissionRequest {
	return &GroupRevokePermissionRequest{
		GroupUID:      strings.TrimSpace(req.GetGroupUid()),
		PermissionUID: strings.TrimSpace(req.GetPermissionUid()),
	}
}

// GetGroupUID returns the group UID.
func (r *GroupRevokePermissionRequest) GetGroupUID() string {
	return r.GroupUID
}

// GetPermissionUID returns the permission UID.
func (r *GroupRevokePermissionRequest) GetPermissionUID() string {
	return r.PermissionUID
}

// GroupListPermissionsRequest represents validated group permission list request data.
type GroupListPermissionsRequest struct {
	GroupUID   string
	Pagination *PaginationRequest
	Filter     *GroupPermissionFilterRequest
}

// GroupListPermissionsRequestFromPb converts a proto ListPermissionsRequest to a GroupListPermissionsRequest.
func GroupListPermissionsRequestFromPb(req *group.ListPermissionsRequest) *GroupListPermissionsRequest {
	payload := &GroupListPermissionsRequest{
		GroupUID: strings.TrimSpace(req.GetGroupUid()),
	}

	if req.Pagination != nil {
		payload.Pagination = PaginationRequestFromPb(req.GetPagination())
	}

	if req.Filter != nil {
		payload.Filter = groupPermissionFilterFromPb(req.GetFilter())
	}

	return payload
}

// GetGroupUID returns the group UID.
func (r *GroupListPermissionsRequest) GetGroupUID() string {
	return r.GroupUID
}

// ToGroupPermissionListParams converts a GroupListPermissionsRequest to domain params.
func (r *GroupListPermissionsRequest) ToGroupPermissionListParams() *param.GroupPermissionListParam {
	var pagination *param.PaginationParam
	if r.Pagination != nil {
		pagination = r.Pagination.ToPaginationParam()
	}

	var filter *param.GroupPermissionListFilterParam
	if r.Filter != nil {
		filter = r.Filter.toGroupPermissionListFilterParam()
	}

	return &param.GroupPermissionListParam{
		Pagination: pagination,
		Filter:     filter,
	}
}

// GroupPermissionFilterRequest represents validated group permission filter request data.
type GroupPermissionFilterRequest struct {
	UIDs           []string
	PermissionUIDs []string
	Resource       *string
	Action         *string
	Query          *string
}

// groupPermissionFilterFromPb converts a proto FilterPermissionRequest to a GroupPermissionFilterRequest.
func groupPermissionFilterFromPb(req *group.FilterPermissionRequest) *GroupPermissionFilterRequest {
	if req == nil {
		return nil
	}

	r := &GroupPermissionFilterRequest{
		UIDs:           req.GetUids(),
		PermissionUIDs: req.GetPermissionUids(),
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

// toGroupPermissionListFilterParam converts a GroupPermissionFilterRequest to domain params.
func (r *GroupPermissionFilterRequest) toGroupPermissionListFilterParam() *param.GroupPermissionListFilterParam {
	return &param.GroupPermissionListFilterParam{
		UIDs:           r.UIDs,
		PermissionUIDs: r.PermissionUIDs,
		Resource:       r.Resource,
		Action:         r.Action,
		Query:          r.Query,
	}
}

// ToGroupCreateParam converts proto CreateRequest directly to domain param.
func ToGroupCreateParam(req *group.CreateRequest) param.GroupCreateParam {
	return param.GroupCreateParam{
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
	}
}

// ToGroupUpdateParam converts proto UpdateRequest directly to domain param.
func ToGroupUpdateParam(req *group.UpdateRequest) param.GroupUpdateParam {
	r := param.GroupUpdateParam{}

	if name := strings.TrimSpace(req.Name); name != "" {
		r.Name = &name
	}
	if description := strings.TrimSpace(req.Description); description != "" {
		r.Description = &description
	}

	return r
}

// ToGroupListFilterParam converts proto ListRequest directly to domain filter param.
func ToGroupListFilterParam(req *group.ListRequest) *param.GroupListFilterParam {
	if req.Filter == nil {
		return &param.GroupListFilterParam{}
	}

	return &param.GroupListFilterParam{
		UIDs:  req.Filter.Uids,
		Name:  req.Filter.Name,
		Query: req.Filter.Query,
	}
}

// ToGroupPermissionListFilterParam converts proto ListPermissionsRequest directly to domain filter param.
func ToGroupPermissionListFilterParam(req *group.ListPermissionsRequest) *param.GroupPermissionListFilterParam {
	if req.Filter == nil {
		return &param.GroupPermissionListFilterParam{}
	}

	return &param.GroupPermissionListFilterParam{
		UIDs:           req.Filter.Uids,
		PermissionUIDs: req.Filter.PermissionUids,
		Resource:       req.Filter.Resource,
		Action:         req.Filter.Action,
		Query:          req.Filter.Query,
	}
}
