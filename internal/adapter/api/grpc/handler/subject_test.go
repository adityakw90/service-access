package handler_test

import (
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

			h := handler.NewSubjectHandler(mockSvc, nil)
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

			h := handler.NewSubjectHandler(mockSvc, nil)
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

			h := handler.NewSubjectHandler(mockSvc, nil)
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
