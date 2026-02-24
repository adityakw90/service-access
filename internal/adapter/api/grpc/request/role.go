package request

import (
	"strings"

	role "github.com/adityakw90/service-access-proto/gen/go/role"
	param "github.com/adityakw90/service-access/internal/core/domain/param"
)

// RoleCreateRequest represents validated role creation request data.
type RoleCreateRequest struct {
	GroupUID    string
	Name        string `validate:"required"`
	Description string
}

// RoleCreateRequestFromPb converts a proto CreateRequest to a RoleCreateRequest.
func RoleCreateRequestFromPb(req *role.CreateRequest) *RoleCreateRequest {
	return &RoleCreateRequest{
		GroupUID:    strings.TrimSpace(req.GetGroupUid()),
		Name:        strings.TrimSpace(req.GetName()),
		Description: strings.TrimSpace(req.GetDescription()),
	}
}

// GetGroupUID returns the group UID for role creation.
func (r *RoleCreateRequest) GetGroupUID() string {
	return r.GroupUID
}

// ToRoleCreateParams converts a RoleCreateRequest to domain params.
// Note: The handler must resolve GroupUID to GroupID before calling the service.
func (r *RoleCreateRequest) ToRoleCreateParams(groupID int64) *param.RoleCreateParam {
	return &param.RoleCreateParam{
		GroupID:     groupID,
		GroupUID:    r.GroupUID,
		Name:        r.Name,
		Description: r.Description,
	}
}

// RoleUpdateRequest represents validated role update request data.
type RoleUpdateRequest struct {
	UID         string
	Name        *string
	Description *string
}

// RoleUpdateRequestFromPb converts a proto UpdateRequest to a RoleUpdateRequest.
func RoleUpdateRequestFromPb(req *role.UpdateRequest) *RoleUpdateRequest {
	r := &RoleUpdateRequest{
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

// ToRoleUpdateParams converts a RoleUpdateRequest to domain params.
func (r *RoleUpdateRequest) ToRoleUpdateParams() *param.RoleUpdateParam {
	return &param.RoleUpdateParam{
		Name:        r.Name,
		Description: r.Description,
	}
}

// GetUID returns the UID for update operations.
func (r *RoleUpdateRequest) GetUID() string {
	return r.UID
}

// RoleGetRequest represents validated role get request data.
type RoleGetRequest struct {
	UID string `validate:"required"`
}

// RoleGetRequestFromPb converts a proto GetRequest to a RoleGetRequest.
func RoleGetRequestFromPb(req *role.GetRequest) *RoleGetRequest {
	return &RoleGetRequest{
		UID: strings.TrimSpace(req.GetUid()),
	}
}

// GetUID returns the UID for get operations.
func (r *RoleGetRequest) GetUID() string {
	return r.UID
}

// RoleDeleteRequest represents validated role delete request data.
type RoleDeleteRequest struct {
	UID string `validate:"required"`
}

// RoleDeleteRequestFromPb converts a proto DeleteRequest to a RoleDeleteRequest.
func RoleDeleteRequestFromPb(req *role.DeleteRequest) *RoleDeleteRequest {
	return &RoleDeleteRequest{
		UID: strings.TrimSpace(req.GetUid()),
	}
}

// GetUID returns the UID for delete operations.
func (r *RoleDeleteRequest) GetUID() string {
	return r.UID
}

// RoleListRequest represents validated role list request data.
type RoleListRequest struct {
	Pagination *PaginationRequest
	Filter     *RoleFilterRequest
}

// RoleListRequestFromPb converts a proto ListRequest to a RoleListRequest.
func RoleListRequestFromPb(req *role.ListRequest) *RoleListRequest {
	payload := &RoleListRequest{}

	if req.Pagination != nil {
		payload.Pagination = PaginationRequestFromPb(req.GetPagination())
	}

	if req.Filter != nil {
		payload.Filter = roleFilterFromPb(req.GetFilter())
	}

	return payload
}

// ToRoleListParams converts a RoleListRequest to domain params.
func (r *RoleListRequest) ToRoleListParams() *param.RoleListParam {
	var pagination *param.PaginationParam
	if r.Pagination != nil {
		pagination = r.Pagination.ToPaginationParam()
	}

	var filter *param.RoleListFilterParam
	if r.Filter != nil {
		filter = r.Filter.toRoleListFilterParam()
	}

	return &param.RoleListParam{
		Pagination: pagination,
		Filter:     filter,
	}
}

// RoleFilterRequest represents validated role filter request data.
type RoleFilterRequest struct {
	UIDs      []string
	GroupUIDs []string
	Name      *string
	Query     *string
}

// roleFilterFromPb converts a proto FilterRequest to a RoleFilterRequest.
func roleFilterFromPb(req *role.FilterRequest) *RoleFilterRequest {
	if req == nil {
		return nil
	}

	r := &RoleFilterRequest{
		UIDs:      req.GetUids(),
		GroupUIDs: req.GetGroupUids(),
	}

	if name := strings.TrimSpace(req.GetName()); name != "" {
		r.Name = &name
	}
	if query := strings.TrimSpace(req.GetQuery()); query != "" {
		r.Query = &query
	}

	return r
}

// toRoleListFilterParam converts a RoleFilterRequest to domain params.
// Note: Domain param only supports filtering by a single group, so we use
// the first GroupUID if multiple are provided. The handler should handle
// multiple group filtering by making multiple calls if needed.
func (r *RoleFilterRequest) toRoleListFilterParam() *param.RoleListFilterParam {
	var groupUID *string
	if len(r.GroupUIDs) > 0 {
		// Domain layer only supports single group filtering
		groupUID = &r.GroupUIDs[0]
	}

	return &param.RoleListFilterParam{
		UIDs:     r.UIDs,
		GroupUID: groupUID,
		Name:     r.Name,
		Query:    r.Query,
	}
}

// RoleUpdatePermissionRequest represents validated role permission update request data.
type RoleUpdatePermissionRequest struct {
	RoleUID             string
	GroupPermissionUIDs []string
}

// RoleUpdatePermissionRequestFromPb converts a proto UpdatePermissionRequest to a RoleUpdatePermissionRequest.
func RoleUpdatePermissionRequestFromPb(req *role.UpdatePermissionRequest) *RoleUpdatePermissionRequest {
	return &RoleUpdatePermissionRequest{
		RoleUID:             strings.TrimSpace(req.GetRoleUid()),
		GroupPermissionUIDs: req.GetGroupPermissionUids(),
	}
}

// GetRoleUID returns the role UID.
func (r *RoleUpdatePermissionRequest) GetRoleUID() string {
	return r.RoleUID
}

// GetGroupPermissionUIDs returns the group permission UIDs.
func (r *RoleUpdatePermissionRequest) GetGroupPermissionUIDs() []string {
	return r.GroupPermissionUIDs
}

// RoleAssignPermissionRequest represents validated role permission assignment request data.
type RoleAssignPermissionRequest struct {
	RoleUID            string
	GroupPermissionUID string
}

// RoleAssignPermissionRequestFromPb converts a proto AssignPermissionRequest to a RoleAssignPermissionRequest.
func RoleAssignPermissionRequestFromPb(req *role.AssignPermissionRequest) *RoleAssignPermissionRequest {
	return &RoleAssignPermissionRequest{
		RoleUID:            strings.TrimSpace(req.GetRoleUid()),
		GroupPermissionUID: strings.TrimSpace(req.GetGroupPermissionUid()),
	}
}

// GetRoleUID returns the role UID.
func (r *RoleAssignPermissionRequest) GetRoleUID() string {
	return r.RoleUID
}

// GetGroupPermissionUID returns the group permission UID.
func (r *RoleAssignPermissionRequest) GetGroupPermissionUID() string {
	return r.GroupPermissionUID
}

// RoleRevokePermissionRequest represents validated role permission revocation request data.
type RoleRevokePermissionRequest struct {
	RoleUID            string
	GroupPermissionUID string
}

// RoleRevokePermissionRequestFromPb converts a proto RevokePermissionRequest to a RoleRevokePermissionRequest.
func RoleRevokePermissionRequestFromPb(req *role.RevokePermissionRequest) *RoleRevokePermissionRequest {
	return &RoleRevokePermissionRequest{
		RoleUID:            strings.TrimSpace(req.GetRoleUid()),
		GroupPermissionUID: strings.TrimSpace(req.GetGroupPermissionUid()),
	}
}

// GetRoleUID returns the role UID.
func (r *RoleRevokePermissionRequest) GetRoleUID() string {
	return r.RoleUID
}

// GetGroupPermissionUID returns the group permission UID.
func (r *RoleRevokePermissionRequest) GetGroupPermissionUID() string {
	return r.GroupPermissionUID
}

// RoleListPermissionsRequest represents validated role permission list request data.
type RoleListPermissionsRequest struct {
	RoleUID    string
	Pagination *PaginationRequest
	Filter     *RolePermissionFilterRequest
}

// RoleListPermissionsRequestFromPb converts a proto ListPermissionsRequest to a RoleListPermissionsRequest.
func RoleListPermissionsRequestFromPb(req *role.ListPermissionsRequest) *RoleListPermissionsRequest {
	payload := &RoleListPermissionsRequest{
		RoleUID: strings.TrimSpace(req.GetRoleUid()),
	}

	if req.Pagination != nil {
		payload.Pagination = PaginationRequestFromPb(req.GetPagination())
	}

	if req.Filter != nil {
		payload.Filter = rolePermissionFilterFromPb(req.GetFilter())
	}

	return payload
}

// GetRoleUID returns the role UID.
func (r *RoleListPermissionsRequest) GetRoleUID() string {
	return r.RoleUID
}

// ToRolePermissionListParams converts a RoleListPermissionsRequest to domain params.
func (r *RoleListPermissionsRequest) ToRolePermissionListParams() *param.RolePermissionListParam {
	var pagination *param.PaginationParam
	if r.Pagination != nil {
		pagination = r.Pagination.ToPaginationParam()
	}

	var filter *param.RolePermissionListFilterParam
	if r.Filter != nil {
		filter = r.Filter.toRolePermissionListFilterParam()
	}

	return &param.RolePermissionListParam{
		Pagination: pagination,
		Filter:     filter,
	}
}

// RolePermissionFilterRequest represents validated role permission filter request data.
type RolePermissionFilterRequest struct {
	PermissionUIDs []string
	Resource       *string
	Action         *string
	Query          *string
}

// rolePermissionFilterFromPb converts a proto FilterPermissionRequest to a RolePermissionFilterRequest.
func rolePermissionFilterFromPb(req *role.FilterPermissionRequest) *RolePermissionFilterRequest {
	if req == nil {
		return nil
	}

	r := &RolePermissionFilterRequest{
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

// toRolePermissionListFilterParam converts a RolePermissionFilterRequest to domain params.
func (r *RolePermissionFilterRequest) toRolePermissionListFilterParam() *param.RolePermissionListFilterParam {
	return &param.RolePermissionListFilterParam{
		PermissionUIDs: r.PermissionUIDs,
		Resource:       r.Resource,
		Action:         r.Action,
		Query:          r.Query,
	}
}

// ToRoleCreateParam converts proto CreateRequest directly to domain param.
func ToRoleCreateParam(req *role.CreateRequest) param.RoleCreateParam {
	return param.RoleCreateParam{
		GroupUID:    strings.TrimSpace(req.GroupUid),
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
	}
}

// ToRoleUpdateParam converts proto UpdateRequest directly to domain param.
func ToRoleUpdateParam(req *role.UpdateRequest) param.RoleUpdateParam {
	r := param.RoleUpdateParam{}

	if name := strings.TrimSpace(req.Name); name != "" {
		r.Name = &name
	}
	if description := strings.TrimSpace(req.Description); description != "" {
		r.Description = &description
	}

	return r
}

// ToRoleListFilterParam converts proto ListRequest directly to domain filter param.
func ToRoleListFilterParam(req *role.ListRequest) *param.RoleListFilterParam {
	if req.Filter == nil {
		return &param.RoleListFilterParam{}
	}

	r := &param.RoleListFilterParam{
		UIDs:  req.Filter.Uids,
		Name:  req.Filter.Name,
		Query: req.Filter.Query,
	}

	// Domain layer only supports single group filtering - use first GroupUID if provided
	if len(req.Filter.GroupUids) > 0 {
		r.GroupUID = &req.Filter.GroupUids[0]
	}

	return r
}

// ToRolePermissionListFilterParam converts proto ListPermissionsRequest directly to domain filter param.
func ToRolePermissionListFilterParam(req *role.ListPermissionsRequest) *param.RolePermissionListFilterParam {
	if req.Filter == nil {
		return &param.RolePermissionListFilterParam{}
	}

	return &param.RolePermissionListFilterParam{
		PermissionUIDs: req.Filter.PermissionUids,
		Resource:       req.Filter.Resource,
		Action:         req.Filter.Action,
		Query:          req.Filter.Query,
	}
}
