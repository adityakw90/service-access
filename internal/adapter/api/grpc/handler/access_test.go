package handler_test

import (
	"context"
	"testing"

	"github.com/adityakw90/service-access/internal/adapter/api/grpc/handler"
	"github.com/adityakw90/service-access/internal/core/domain/model"
	accessmocks "github.com/adityakw90/service-access/test/mocks/service"
	subjectmocks "github.com/adityakw90/service-access/test/mocks/service"
	"github.com/adityakw90/service-access-proto/gen/go/access"
	"github.com/stretchr/testify/mock"
)

func TestAccessHandler_CheckAccess(t *testing.T) {
	tests := []struct {
		name       string
		req        *access.CheckAccessRequest
		setup      func(*accessmocks.MockAccessService)
		wantAllowed bool
		wantErr    bool
	}{
		{
			name: "Access Granted",
			req: &access.CheckAccessRequest{
				SubjectId:   "user-123",
				SubjectType: "user",
				Resource:    "invoices",
				Action:      "read",
			},
			setup: func(m *accessmocks.MockAccessService) {
				m.EXPECT().Check(mock.Anything, "user-123", "user", "invoices", "read").Return(
					true, "Access granted via admin role", nil,
				)
			},
			wantAllowed: true,
		},
		{
			name: "Access Denied",
			req: &access.CheckAccessRequest{
				SubjectId:   "user-456",
				SubjectType: "user",
				Resource:    "admin",
				Action:      "delete",
			},
			setup: func(m *accessmocks.MockAccessService) {
				m.EXPECT().Check(mock.Anything, "user-456", "user", "admin", "delete").Return(
					false, "No matching permissions", nil,
				)
			},
			wantAllowed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(accessmocks.MockAccessService)
			if tt.setup != nil {
				tt.setup(mockSvc)
			}

			h := handler.NewAccessHandler(mockSvc, nil, nil)
			got, err := h.CheckAccess(context.Background(), tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("CheckAccess() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got.Allowed != tt.wantAllowed {
				t.Errorf("CheckAccess() allowed = %v, want %v", got.Allowed, tt.wantAllowed)
			}
		})
	}
}

func TestAccessHandler_AssignRole(t *testing.T) {
	tests := []struct {
		name    string
		req     *access.AssignRoleRequest
		setup   func(*subjectmocks.MockSubjectService)
		wantErr bool
	}{
		{
			name: "Happy Path",
			req: &access.AssignRoleRequest{
				SubjectId:   "user-123",
				SubjectType: "user",
				RoleUid:     "role-123",
			},
			setup: func(m *subjectmocks.MockSubjectService) {
				m.EXPECT().Assign(mock.Anything, "user-123", "user", "role-123").Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(subjectmocks.MockSubjectService)
			if tt.setup != nil {
				tt.setup(mockSvc)
			}

			h := handler.NewAccessHandler(nil, mockSvc, nil)
			_, err := h.AssignRole(context.Background(), tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("AssignRole() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAccessHandler_ListSubjectRoles(t *testing.T) {
	tests := []struct {
		name    string
		req     *access.ListSubjectRolesRequest
		setup   func(*subjectmocks.MockSubjectService)
		wantErr bool
	}{
		{
			name: "Happy Path",
			req:  &access.ListSubjectRolesRequest{},
			setup: func(m *subjectmocks.MockSubjectService) {
				m.EXPECT().List(mock.Anything, mock.Anything, mock.Anything).Return(
					&model.SubjectRoles{
						Items: []model.SubjectRole{
							{SubjectID: "user-123", RoleUID: "role-1"},
							{SubjectID: "user-123", RoleUID: "role-2"},
						},
						Meta: model.Meta{Page: 1, Limit: 10, Total: 2, Pages: 1},
					},
					nil,
				)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(subjectmocks.MockSubjectService)
			if tt.setup != nil {
				tt.setup(mockSvc)
			}

			h := handler.NewAccessHandler(nil, mockSvc, nil)
			got, err := h.ListSubjectRoles(context.Background(), tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("ListSubjectRoles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(got.Items) != 2 {
				t.Errorf("ListSubjectRoles() got %d items, want 2", len(got.Items))
			}
		})
	}
}
