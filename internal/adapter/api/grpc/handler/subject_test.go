package handler_test

import (
	"github.com/adityakw90/service-access/internal/adapter/api/grpc/validator"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/adityakw90/service-access/internal/adapter/api/grpc/handler"
	"github.com/adityakw90/service-access/internal/core/domain/model"
	subjectmocks "github.com/adityakw90/service-access/test/mocks/service"
	"github.com/adityakw90/service-access-proto/gen/go/common"
	"github.com/adityakw90/service-access-proto/gen/go/subject"
	"github.com/stretchr/testify/mock"
)

// ptr returns a pointer to the given string value (helper for optional proto fields)
func ptr(s string) *string {
	return &s
}

func TestSubjectHandler_AssignRole(t *testing.T) {
	tests := []struct {
		name    string
		req     *subject.AssignRoleRequest
		setup   func(*subjectmocks.MockSubjectService)
		wantErr bool
	}{
		{
			name: "Happy Path - User Role Assignment",
			req: &subject.AssignRoleRequest{
				SubjectId:   "user-123",
				SubjectType: "user",
				RoleUid:     "role-123",
			},
			setup: func(m *subjectmocks.MockSubjectService) {
				m.EXPECT().Assign(mock.Anything, "user-123", "user", "role-123").Return(nil)
			},
			wantErr: false,
		},
		{
			name: "Happy Path - Service Account Assignment",
			req: &subject.AssignRoleRequest{
				SubjectId:   "service-456",
				SubjectType: "service_account",
				RoleUid:     "role-789",
			},
			setup: func(m *subjectmocks.MockSubjectService) {
				m.EXPECT().Assign(mock.Anything, "service-456", "service_account", "role-789").Return(nil)
			},
			wantErr: false,
		},
		{
			name: "Service Error - Role Not Found",
			req: &subject.AssignRoleRequest{
				SubjectId:   "user-123",
				SubjectType: "user",
				RoleUid:     "nonexistent-role",
			},
			setup: func(m *subjectmocks.MockSubjectService) {
				m.EXPECT().Assign(mock.Anything, "user-123", "user", "nonexistent-role").Return(
					errors.New("role not found"),
				)
			},
			wantErr: true,
		},
		{
			name: "Service Error - Subject Not Found",
			req: &subject.AssignRoleRequest{
				SubjectId:   "nonexistent-user",
				SubjectType: "user",
				RoleUid:     "role-123",
			},
			setup: func(m *subjectmocks.MockSubjectService) {
				m.EXPECT().Assign(mock.Anything, "nonexistent-user", "user", "role-123").Return(
					errors.New("subject not found"),
				)
			},
			wantErr: true,
		},
		{
			name: "Service Error - Already Assigned",
			req: &subject.AssignRoleRequest{
				SubjectId:   "user-123",
				SubjectType: "user",
				RoleUid:     "role-123",
			},
			setup: func(m *subjectmocks.MockSubjectService) {
				m.EXPECT().Assign(mock.Anything, "user-123", "user", "role-123").Return(
					errors.New("role already assigned to subject"),
				)
			},
			wantErr: true,
		},
		{
			name: "Service Error - Database Error",
			req: &subject.AssignRoleRequest{
				SubjectId:   "user-123",
				SubjectType: "user",
				RoleUid:     "role-123",
			},
			setup: func(m *subjectmocks.MockSubjectService) {
				m.EXPECT().Assign(mock.Anything, "user-123", "user", "role-123").Return(
					errors.New("database connection failed"),
				)
			},
			wantErr: true,
		},
		{
			name: "Empty Subject ID",
			req: &subject.AssignRoleRequest{
				SubjectId:   "",
				SubjectType: "user",
				RoleUid:     "role-123",
			},
			setup: func(m *subjectmocks.MockSubjectService) {
				// Handler may validate before calling service, or service may validate
				// Either way, we expect an error
				m.EXPECT().Assign(mock.Anything, "", "user", "role-123").Return(
					errors.New("subject_id cannot be empty"),
				).Maybe()
			},
			wantErr: true,
		},
		{
			name: "Empty Subject Type",
			req: &subject.AssignRoleRequest{
				SubjectId:   "user-123",
				SubjectType: "",
				RoleUid:     "role-123",
			},
			setup: func(m *subjectmocks.MockSubjectService) {
				m.EXPECT().Assign(mock.Anything, "user-123", "", "role-123").Return(
					errors.New("subject_type cannot be empty"),
				).Maybe()
			},
			wantErr: true,
		},
		{
			name: "Empty Role UID",
			req: &subject.AssignRoleRequest{
				SubjectId:   "user-123",
				SubjectType: "user",
				RoleUid:     "",
			},
			setup: func(m *subjectmocks.MockSubjectService) {
				m.EXPECT().Assign(mock.Anything, "user-123", "user", "").Return(
					errors.New("role_uid cannot be empty"),
				).Maybe()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(subjectmocks.MockSubjectService)
			if tt.setup != nil {
				tt.setup(mockSvc)
			}

			h := handler.NewSubjectHandler(mockSvc, validator.New())
			got, err := h.AssignRole(context.Background(), tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("AssignRole() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// On success, verify response
			if !tt.wantErr {
				if got == nil {
					t.Error("AssignRole() expected non-nil response on success")
				} else if !got.Success {
					t.Error("AssignRole() expected Success = true on success")
				}
			}
		})
	}
}

func TestSubjectHandler_RevokeRole(t *testing.T) {
	tests := []struct {
		name    string
		req     *subject.RevokeRoleRequest
		setup   func(*subjectmocks.MockSubjectService)
		wantErr bool
	}{
		{
			name: "Happy Path - User Role Revocation",
			req: &subject.RevokeRoleRequest{
				SubjectId:   "user-123",
				SubjectType: "user",
				RoleUid:     "role-123",
			},
			setup: func(m *subjectmocks.MockSubjectService) {
				m.EXPECT().Revoke(mock.Anything, "user-123", "user", "role-123").Return(nil)
			},
			wantErr: false,
		},
		{
			name: "Happy Path - Service Account Revocation",
			req: &subject.RevokeRoleRequest{
				SubjectId:   "service-456",
				SubjectType: "service_account",
				RoleUid:     "role-789",
			},
			setup: func(m *subjectmocks.MockSubjectService) {
				m.EXPECT().Revoke(mock.Anything, "service-456", "service_account", "role-789").Return(nil)
			},
			wantErr: false,
		},
		{
			name: "Service Error - Role Not Assigned",
			req: &subject.RevokeRoleRequest{
				SubjectId:   "user-123",
				SubjectType: "user",
				RoleUid:     "role-123",
			},
			setup: func(m *subjectmocks.MockSubjectService) {
				m.EXPECT().Revoke(mock.Anything, "user-123", "user", "role-123").Return(
					errors.New("role not assigned to subject"),
				)
			},
			wantErr: true,
		},
		{
			name: "Service Error - Subject Not Found",
			req: &subject.RevokeRoleRequest{
				SubjectId:   "nonexistent-user",
				SubjectType: "user",
				RoleUid:     "role-123",
			},
			setup: func(m *subjectmocks.MockSubjectService) {
				m.EXPECT().Revoke(mock.Anything, "nonexistent-user", "user", "role-123").Return(
					errors.New("subject not found"),
				)
			},
			wantErr: true,
		},
		{
			name: "Service Error - Role Not Found",
			req: &subject.RevokeRoleRequest{
				SubjectId:   "user-123",
				SubjectType: "user",
				RoleUid:     "nonexistent-role",
			},
			setup: func(m *subjectmocks.MockSubjectService) {
				m.EXPECT().Revoke(mock.Anything, "user-123", "user", "nonexistent-role").Return(
					errors.New("role not found"),
				)
			},
			wantErr: true,
		},
		{
			name: "Service Error - Database Error",
			req: &subject.RevokeRoleRequest{
				SubjectId:   "user-123",
				SubjectType: "user",
				RoleUid:     "role-123",
			},
			setup: func(m *subjectmocks.MockSubjectService) {
				m.EXPECT().Revoke(mock.Anything, "user-123", "user", "role-123").Return(
					errors.New("database connection failed"),
				)
			},
			wantErr: true,
		},
		{
			name: "Empty Subject ID",
			req: &subject.RevokeRoleRequest{
				SubjectId:   "",
				SubjectType: "user",
				RoleUid:     "role-123",
			},
			setup: func(m *subjectmocks.MockSubjectService) {
				m.EXPECT().Revoke(mock.Anything, "", "user", "role-123").Return(
					errors.New("subject_id cannot be empty"),
				).Maybe()
			},
			wantErr: true,
		},
		{
			name: "Empty Subject Type",
			req: &subject.RevokeRoleRequest{
				SubjectId:   "user-123",
				SubjectType: "",
				RoleUid:     "role-123",
			},
			setup: func(m *subjectmocks.MockSubjectService) {
				m.EXPECT().Revoke(mock.Anything, "user-123", "", "role-123").Return(
					errors.New("subject_type cannot be empty"),
				).Maybe()
			},
			wantErr: true,
		},
		{
			name: "Empty Role UID",
			req: &subject.RevokeRoleRequest{
				SubjectId:   "user-123",
				SubjectType: "user",
				RoleUid:     "",
			},
			setup: func(m *subjectmocks.MockSubjectService) {
				m.EXPECT().Revoke(mock.Anything, "user-123", "user", "").Return(
					errors.New("role_uid cannot be empty"),
				).Maybe()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(subjectmocks.MockSubjectService)
			if tt.setup != nil {
				tt.setup(mockSvc)
			}

			h := handler.NewSubjectHandler(mockSvc, validator.New())
			got, err := h.RevokeRole(context.Background(), tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("RevokeRole() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// On success, verify response
			if !tt.wantErr {
				if got == nil {
					t.Error("RevokeRole() expected non-nil response on success")
				} else if !got.Success {
					t.Error("RevokeRole() expected Success = true on success")
				}
			}
		})
	}
}

func TestSubjectHandler_List(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		req         *subject.ListRequest
		setup       func(*subjectmocks.MockSubjectService)
		wantErr     bool
		validateResp func(*testing.T, *subject.ListResponse)
	}{
		{
			name: "Happy Path - List All",
			req:  &subject.ListRequest{},
			setup: func(m *subjectmocks.MockSubjectService) {
				m.EXPECT().List(mock.Anything, mock.Anything, mock.Anything).Return(
					&model.SubjectRoles{
						Items: []model.SubjectRole{
							{SubjectID: "user-123", RoleUID: "role-1", AssignedAt: now},
							{SubjectID: "user-123", RoleUID: "role-2", AssignedAt: now},
						},
						Meta: model.Meta{Page: 1, Limit: 10, Total: 2, Pages: 1},
					},
					nil,
				)
			},
			wantErr: false,
			validateResp: func(t *testing.T, got *subject.ListResponse) {
				if len(got.Items) != 2 {
					t.Errorf("List() got %d items, want 2", len(got.Items))
				}
				if got.Meta == nil {
					t.Error("List() expected non-nil Meta")
				} else {
					if got.Meta.Page != 1 {
						t.Errorf("List() Meta.Page = %d, want 1", got.Meta.Page)
					}
					if got.Meta.Total != int64(2) {
						t.Errorf("List() Meta.Total = %d, want 2", got.Meta.Total)
					}
				}
			},
		},
		{
			name: "Happy Path - Empty List",
			req:  &subject.ListRequest{},
			setup: func(m *subjectmocks.MockSubjectService) {
				m.EXPECT().List(mock.Anything, mock.Anything, mock.Anything).Return(
					&model.SubjectRoles{
						Items: []model.SubjectRole{},
						Meta:  model.Meta{Page: 1, Limit: 10, Total: 0, Pages: 0},
					},
					nil,
				)
			},
			wantErr: false,
			validateResp: func(t *testing.T, got *subject.ListResponse) {
				if len(got.Items) != 0 {
					t.Errorf("List() got %d items, want 0", len(got.Items))
				}
			},
		},
		{
			name: "Happy Path - With Pagination",
			req: &subject.ListRequest{
				Pagination: &common.Pagination{
					Page:  2,
					Limit: 5,
				},
			},
			setup: func(m *subjectmocks.MockSubjectService) {
				m.EXPECT().List(mock.Anything, mock.Anything, mock.Anything).Return(
					&model.SubjectRoles{
						Items: []model.SubjectRole{
							{SubjectID: "user-456", RoleUID: "role-3", AssignedAt: now},
						},
						Meta: model.Meta{Page: 2, Limit: 5, Total: 12, Pages: 3},
					},
					nil,
				)
			},
			wantErr: false,
			validateResp: func(t *testing.T, got *subject.ListResponse) {
				if len(got.Items) != 1 {
					t.Errorf("List() got %d items, want 1", len(got.Items))
				}
				if got.Meta == nil {
					t.Error("List() expected non-nil Meta")
				} else {
					if got.Meta.Page != 2 {
						t.Errorf("List() Meta.Page = %d, want 2", got.Meta.Page)
					}
					if got.Meta.Total != int64(12) {
						t.Errorf("List() Meta.Total = %d, want 12", got.Meta.Total)
					}
				}
			},
		},
		{
			name: "Happy Path - With Filter",
			req: &subject.ListRequest{
				Filter: &subject.FilterRequest{
					SubjectId: ptr("user-123"),
				},
			},
			setup: func(m *subjectmocks.MockSubjectService) {
				m.EXPECT().List(mock.Anything, mock.Anything, mock.Anything).Return(
					&model.SubjectRoles{
						Items: []model.SubjectRole{
							{SubjectID: "user-123", RoleUID: "role-1", AssignedAt: now},
						},
						Meta: model.Meta{Page: 1, Limit: 10, Total: 1, Pages: 1},
					},
					nil,
				)
			},
			wantErr: false,
			validateResp: func(t *testing.T, got *subject.ListResponse) {
				if len(got.Items) != 1 {
					t.Errorf("List() got %d items, want 1", len(got.Items))
				}
			},
		},
		{
			name: "Service Error - Database Error",
			req:  &subject.ListRequest{},
			setup: func(m *subjectmocks.MockSubjectService) {
				m.EXPECT().List(mock.Anything, mock.Anything, mock.Anything).Return(
					nil,
					errors.New("database connection failed"),
				)
			},
			wantErr: true,
		},
		{
			name: "Service Error - Invalid Pagination",
			req: &subject.ListRequest{
				Pagination: &common.Pagination{
					Page:  0, // Invalid page
					Limit: 10,
				},
			},
			setup: func(m *subjectmocks.MockSubjectService) {
				m.EXPECT().List(mock.Anything, mock.Anything, mock.Anything).Return(
					nil,
					errors.New("invalid pagination parameters"),
				)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(subjectmocks.MockSubjectService)
			if tt.setup != nil {
				tt.setup(mockSvc)
			}

			h := handler.NewSubjectHandler(mockSvc, validator.New())
			got, err := h.List(context.Background(), tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// On success, validate response
			if !tt.wantErr && tt.validateResp != nil {
				tt.validateResp(t, got)
			}
		})
	}
}

func TestSubjectHandler_Get(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		req         *subject.GetSubjectRequest
		setup       func(*subjectmocks.MockSubjectService)
		wantErr     bool
		validateResp func(*testing.T, *subject.GetSubjectResponse)
	}{
		{
			name: "Happy Path - Get Full Profile",
			req: &subject.GetSubjectRequest{
				SubjectId:   "user-123",
				SubjectType: "user",
			},
			setup: func(m *subjectmocks.MockSubjectService) {
				m.EXPECT().GetFullProfile(mock.Anything, "user-123", "user").Return(
					&model.SubjectProfile{
						Groups: []model.Group{
							{UID: "group-1", Name: "Admin", CreatedAt: now, UpdatedAt: now},
							{UID: "group-2", Name: "Editors", CreatedAt: now, UpdatedAt: now},
						},
						Roles: []model.Role{
							{UID: "role-1", GroupUID: "group-1", Name: "SuperAdmin", CreatedAt: now, UpdatedAt: now},
							{UID: "role-2", GroupUID: "group-2", Name: "Editor", CreatedAt: now, UpdatedAt: now},
						},
						Permissions: []model.Permission{
							{UID: "perm-1", Resource: "article", Action: "create", CreatedAt: now, UpdatedAt: now},
							{UID: "perm-2", Resource: "article", Action: "delete", CreatedAt: now, UpdatedAt: now},
						},
					},
					nil,
				)
			},
			wantErr: false,
			validateResp: func(t *testing.T, got *subject.GetSubjectResponse) {
				if got.TotalGroup != 2 {
					t.Errorf("Get() TotalGroup = %d, want 2", got.TotalGroup)
				}
				if got.TotalRole != 2 {
					t.Errorf("Get() TotalRole = %d, want 2", got.TotalRole)
				}
				if got.TotalPermission != 2 {
					t.Errorf("Get() TotalPermission = %d, want 2", got.TotalPermission)
				}
				if len(got.Groups) != 2 {
					t.Errorf("Get() groups count = %d, want 2", len(got.Groups))
				}
				if len(got.Roles) != 2 {
					t.Errorf("Get() roles count = %d, want 2", len(got.Roles))
				}
				if len(got.Permissions) != 2 {
					t.Errorf("Get() permissions count = %d, want 2", len(got.Permissions))
				}
			},
		},
		{
			name: "Happy Path - Empty Profile",
			req: &subject.GetSubjectRequest{
				SubjectId:   "user-999",
				SubjectType: "user",
			},
			setup: func(m *subjectmocks.MockSubjectService) {
				m.EXPECT().GetFullProfile(mock.Anything, "user-999", "user").Return(
					&model.SubjectProfile{
						Groups:      []model.Group{},
						Roles:       []model.Role{},
						Permissions: []model.Permission{},
					},
					nil,
				)
			},
			wantErr: false,
			validateResp: func(t *testing.T, got *subject.GetSubjectResponse) {
				if got.TotalGroup != 0 {
					t.Errorf("Get() TotalGroup = %d, want 0", got.TotalGroup)
				}
				if got.TotalRole != 0 {
					t.Errorf("Get() TotalRole = %d, want 0", got.TotalRole)
				}
				if got.TotalPermission != 0 {
					t.Errorf("Get() TotalPermission = %d, want 0", got.TotalPermission)
				}
			},
		},
		{
			name: "Missing Subject ID",
			req: &subject.GetSubjectRequest{
				SubjectId:   "",
				SubjectType: "user",
			},
			setup: func(m *subjectmocks.MockSubjectService) {
				// Validation should prevent service call
				m.EXPECT().GetFullProfile(mock.Anything, mock.Anything, mock.Anything).Return(
					nil,
					errors.New("should not be called"),
				).Maybe()
			},
			wantErr: true,
		},
		{
			name: "Missing Subject Type",
			req: &subject.GetSubjectRequest{
				SubjectId:   "user-123",
				SubjectType: "",
			},
			setup: func(m *subjectmocks.MockSubjectService) {
				// Validation should prevent service call
				m.EXPECT().GetFullProfile(mock.Anything, mock.Anything, mock.Anything).Return(
					nil,
					errors.New("should not be called"),
				).Maybe()
			},
			wantErr: true,
		},
		{
			name: "Service Error - Subject Not Found",
			req: &subject.GetSubjectRequest{
				SubjectId:   "nonexistent",
				SubjectType: "user",
			},
			setup: func(m *subjectmocks.MockSubjectService) {
				m.EXPECT().GetFullProfile(mock.Anything, "nonexistent", "user").Return(
					nil,
					errors.New("subject not found"),
				)
			},
			wantErr: true,
		},
		{
			name: "Service Error - Database Error",
			req: &subject.GetSubjectRequest{
				SubjectId:   "user-123",
				SubjectType: "user",
			},
			setup: func(m *subjectmocks.MockSubjectService) {
				m.EXPECT().GetFullProfile(mock.Anything, "user-123", "user").Return(
					nil,
					errors.New("database connection failed"),
				)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(subjectmocks.MockSubjectService)
			if tt.setup != nil {
				tt.setup(mockSvc)
			}

			h := handler.NewSubjectHandler(mockSvc, validator.New())
			got, err := h.Get(context.Background(), tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validateResp != nil {
				tt.validateResp(t, got)
			}
		})
	}
}

func TestSubjectHandler_ListGroup(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		req         *subject.GetSubjectRequest
		setup       func(*subjectmocks.MockSubjectService)
		wantErr     bool
		validateResp func(*testing.T, *subject.ListGroupResponse)
	}{
		{
			name: "Happy Path - List Groups",
			req: &subject.GetSubjectRequest{
				SubjectId:   "user-123",
				SubjectType: "user",
			},
			setup: func(m *subjectmocks.MockSubjectService) {
				m.EXPECT().GetGroups(mock.Anything, "user-123", "user").Return(
					[]model.Group{
						{UID: "group-1", Name: "Admin", CreatedAt: now, UpdatedAt: now},
						{UID: "group-2", Name: "Editors", CreatedAt: now, UpdatedAt: now},
					},
					nil,
				)
			},
			wantErr: false,
			validateResp: func(t *testing.T, got *subject.ListGroupResponse) {
				if got.Total != 2 {
					t.Errorf("ListGroup() Total = %d, want 2", got.Total)
				}
				if len(got.Groups) != 2 {
					t.Errorf("ListGroup() groups count = %d, want 2", len(got.Groups))
				}
			},
		},
		{
			name: "Happy Path - No Groups",
			req: &subject.GetSubjectRequest{
				SubjectId:   "user-999",
				SubjectType: "user",
			},
			setup: func(m *subjectmocks.MockSubjectService) {
				m.EXPECT().GetGroups(mock.Anything, "user-999", "user").Return(
					[]model.Group{},
					nil,
				)
			},
			wantErr: false,
			validateResp: func(t *testing.T, got *subject.ListGroupResponse) {
				if got.Total != 0 {
					t.Errorf("ListGroup() Total = %d, want 0", got.Total)
				}
				if len(got.Groups) != 0 {
					t.Errorf("ListGroup() groups count = %d, want 0", len(got.Groups))
				}
			},
		},
		{
			name: "Missing Subject ID",
			req: &subject.GetSubjectRequest{
				SubjectId:   "",
				SubjectType: "user",
			},
			setup: func(m *subjectmocks.MockSubjectService) {
				m.EXPECT().GetGroups(mock.Anything, "", "user").Return(
					nil,
					errors.New("should not be called"),
				).Maybe()
			},
			wantErr: true,
		},
		{
			name: "Missing Subject Type",
			req: &subject.GetSubjectRequest{
				SubjectId:   "user-123",
				SubjectType: "",
			},
			setup: func(m *subjectmocks.MockSubjectService) {
				m.EXPECT().GetGroups(mock.Anything, "user-123", "").Return(
					nil,
					errors.New("should not be called"),
				).Maybe()
			},
			wantErr: true,
		},
		{
			name: "Service Error - Database Error",
			req: &subject.GetSubjectRequest{
				SubjectId:   "user-123",
				SubjectType: "user",
			},
			setup: func(m *subjectmocks.MockSubjectService) {
				m.EXPECT().GetGroups(mock.Anything, "user-123", "user").Return(
					nil,
					errors.New("database connection failed"),
				)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(subjectmocks.MockSubjectService)
			if tt.setup != nil {
				tt.setup(mockSvc)
			}

			h := handler.NewSubjectHandler(mockSvc, validator.New())
			got, err := h.ListGroup(context.Background(), tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("ListGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validateResp != nil {
				tt.validateResp(t, got)
			}
		})
	}
}

func TestSubjectHandler_ListRole(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		req         *subject.GetSubjectRequest
		setup       func(*subjectmocks.MockSubjectService)
		wantErr     bool
		validateResp func(*testing.T, *subject.ListRoleResponse)
	}{
		{
			name: "Happy Path - List Roles",
			req: &subject.GetSubjectRequest{
				SubjectId:   "user-123",
				SubjectType: "user",
			},
			setup: func(m *subjectmocks.MockSubjectService) {
				m.EXPECT().GetRoles(mock.Anything, "user-123", "user").Return(
					[]model.Role{
						{UID: "role-1", GroupUID: "group-1", Name: "SuperAdmin", CreatedAt: now, UpdatedAt: now},
						{UID: "role-2", GroupUID: "group-2", Name: "Editor", CreatedAt: now, UpdatedAt: now},
					},
					nil,
				)
			},
			wantErr: false,
			validateResp: func(t *testing.T, got *subject.ListRoleResponse) {
				if got.Total != 2 {
					t.Errorf("ListRole() Total = %d, want 2", got.Total)
				}
				if len(got.Roles) != 2 {
					t.Errorf("ListRole() roles count = %d, want 2", len(got.Roles))
				}
			},
		},
		{
			name: "Happy Path - No Roles",
			req: &subject.GetSubjectRequest{
				SubjectId:   "user-999",
				SubjectType: "user",
			},
			setup: func(m *subjectmocks.MockSubjectService) {
				m.EXPECT().GetRoles(mock.Anything, "user-999", "user").Return(
					[]model.Role{},
					nil,
				)
			},
			wantErr: false,
			validateResp: func(t *testing.T, got *subject.ListRoleResponse) {
				if got.Total != 0 {
					t.Errorf("ListRole() Total = %d, want 0", got.Total)
				}
				if len(got.Roles) != 0 {
					t.Errorf("ListRole() roles count = %d, want 0", len(got.Roles))
				}
			},
		},
		{
			name: "Missing Subject ID",
			req: &subject.GetSubjectRequest{
				SubjectId:   "",
				SubjectType: "user",
			},
			setup: func(m *subjectmocks.MockSubjectService) {
				m.EXPECT().GetRoles(mock.Anything, "", "user").Return(
					nil,
					errors.New("should not be called"),
				).Maybe()
			},
			wantErr: true,
		},
		{
			name: "Missing Subject Type",
			req: &subject.GetSubjectRequest{
				SubjectId:   "user-123",
				SubjectType: "",
			},
			setup: func(m *subjectmocks.MockSubjectService) {
				m.EXPECT().GetRoles(mock.Anything, "user-123", "").Return(
					nil,
					errors.New("should not be called"),
				).Maybe()
			},
			wantErr: true,
		},
		{
			name: "Service Error - Database Error",
			req: &subject.GetSubjectRequest{
				SubjectId:   "user-123",
				SubjectType: "user",
			},
			setup: func(m *subjectmocks.MockSubjectService) {
				m.EXPECT().GetRoles(mock.Anything, "user-123", "user").Return(
					nil,
					errors.New("database connection failed"),
				)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(subjectmocks.MockSubjectService)
			if tt.setup != nil {
				tt.setup(mockSvc)
			}

			h := handler.NewSubjectHandler(mockSvc, validator.New())
			got, err := h.ListRole(context.Background(), tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("ListRole() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validateResp != nil {
				tt.validateResp(t, got)
			}
		})
	}
}

func TestSubjectHandler_ListPermission(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		req         *subject.GetSubjectRequest
		setup       func(*subjectmocks.MockSubjectService)
		wantErr     bool
		validateResp func(*testing.T, *subject.ListPermissionResponse)
	}{
		{
			name: "Happy Path - List Permissions",
			req: &subject.GetSubjectRequest{
				SubjectId:   "user-123",
				SubjectType: "user",
			},
			setup: func(m *subjectmocks.MockSubjectService) {
				m.EXPECT().GetPermissions(mock.Anything, "user-123", "user").Return(
					[]model.Permission{
						{UID: "perm-1", Resource: "article", Action: "create", CreatedAt: now, UpdatedAt: now},
						{UID: "perm-2", Resource: "article", Action: "delete", CreatedAt: now, UpdatedAt: now},
					},
					nil,
				)
			},
			wantErr: false,
			validateResp: func(t *testing.T, got *subject.ListPermissionResponse) {
				if got.Total != 2 {
					t.Errorf("ListPermission() Total = %d, want 2", got.Total)
				}
				if len(got.Permissions) != 2 {
					t.Errorf("ListPermission() permissions count = %d, want 2", len(got.Permissions))
				}
			},
		},
		{
			name: "Happy Path - No Permissions",
			req: &subject.GetSubjectRequest{
				SubjectId:   "user-999",
				SubjectType: "user",
			},
			setup: func(m *subjectmocks.MockSubjectService) {
				m.EXPECT().GetPermissions(mock.Anything, "user-999", "user").Return(
					[]model.Permission{},
					nil,
				)
			},
			wantErr: false,
			validateResp: func(t *testing.T, got *subject.ListPermissionResponse) {
				if got.Total != 0 {
					t.Errorf("ListPermission() Total = %d, want 0", got.Total)
				}
				if len(got.Permissions) != 0 {
					t.Errorf("ListPermission() permissions count = %d, want 0", len(got.Permissions))
				}
			},
		},
		{
			name: "Missing Subject ID",
			req: &subject.GetSubjectRequest{
				SubjectId:   "",
				SubjectType: "user",
			},
			setup: func(m *subjectmocks.MockSubjectService) {
				m.EXPECT().GetPermissions(mock.Anything, "", "user").Return(
					nil,
					errors.New("should not be called"),
				).Maybe()
			},
			wantErr: true,
		},
		{
			name: "Missing Subject Type",
			req: &subject.GetSubjectRequest{
				SubjectId:   "user-123",
				SubjectType: "",
			},
			setup: func(m *subjectmocks.MockSubjectService) {
				m.EXPECT().GetPermissions(mock.Anything, "user-123", "").Return(
					nil,
					errors.New("should not be called"),
				).Maybe()
			},
			wantErr: true,
		},
		{
			name: "Service Error - Database Error",
			req: &subject.GetSubjectRequest{
				SubjectId:   "user-123",
				SubjectType: "user",
			},
			setup: func(m *subjectmocks.MockSubjectService) {
				m.EXPECT().GetPermissions(mock.Anything, "user-123", "user").Return(
					nil,
					errors.New("database connection failed"),
				)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(subjectmocks.MockSubjectService)
			if tt.setup != nil {
				tt.setup(mockSvc)
			}

			h := handler.NewSubjectHandler(mockSvc, validator.New())
			got, err := h.ListPermission(context.Background(), tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("ListPermission() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validateResp != nil {
				tt.validateResp(t, got)
			}
		})
	}
}
