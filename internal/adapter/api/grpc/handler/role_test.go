package handler_test

import (
	"context"
	"testing"

	"github.com/adityakw90/service-access/internal/adapter/api/grpc/handler"
	"github.com/adityakw90/service-access/internal/core/domain/model"
	domainErrors "github.com/adityakw90/service-access/internal/core/domain/errors"
	servicemocks "github.com/adityakw90/service-access/test/mocks/service"
	"github.com/adityakw90/service-access-proto/gen/go/role"
	"github.com/stretchr/testify/mock"
)

func TestRoleHandler_Create(t *testing.T) {
	tests := []struct {
		name    string
		req     *role.CreateRequest
		setup   func(*servicemocks.MockRoleService)
		want    *role.CreateResponse
		wantErr bool
	}{
		{
			name: "Happy Path",
			req: &role.CreateRequest{
				GroupUid:    "group-123",
				Name:        "admin",
				Description: "Admin role",
			},
			setup: func(m *servicemocks.MockRoleService) {
				m.EXPECT().Create(mock.Anything, mock.Anything).Return(
					&model.Role{UID: "role-123", GroupUID: "group-123", Name: "admin"}, nil,
				)
			},
			want: &role.CreateResponse{Uid: "role-123"},
		},
		{
			name: "Not Found - Group doesn't exist",
			req: &role.CreateRequest{
				GroupUid: "unknown-group",
				Name:     "admin",
			},
			setup: func(m *servicemocks.MockRoleService) {
				m.EXPECT().Create(mock.Anything, mock.Anything).Return(nil, domainErrors.ErrNotFound)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(servicemocks.MockRoleService)
			if tt.setup != nil {
				tt.setup(mockSvc)
			}

			h := handler.NewRoleHandler(mockSvc, nil)
			got, err := h.Create(context.Background(), tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got.Uid != tt.want.Uid {
				t.Errorf("Create() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoleHandler_List(t *testing.T) {
	tests := []struct {
		name    string
		req     *role.ListRequest
		setup   func(*servicemocks.MockRoleService)
		wantErr bool
	}{
		{
			name: "Happy Path",
			req:  &role.ListRequest{},
			setup: func(m *servicemocks.MockRoleService) {
				m.EXPECT().List(mock.Anything, mock.Anything, mock.Anything).Return(
					&model.Roles{
						Items: []model.Role{{UID: "role-1"}, {UID: "role-2"}},
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
			mockSvc := new(servicemocks.MockRoleService)
			if tt.setup != nil {
				tt.setup(mockSvc)
			}

			h := handler.NewRoleHandler(mockSvc, nil)
			got, err := h.List(context.Background(), tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(got.Items) != 2 {
				t.Errorf("List() got %d items, want 2", len(got.Items))
			}
		})
	}
}

func TestRoleHandler_AssignPermission(t *testing.T) {
	tests := []struct {
		name    string
		req     *role.AssignPermissionRequest
		setup   func(*servicemocks.MockRoleService)
		wantErr bool
	}{
		{
			name: "Happy Path",
			req: &role.AssignPermissionRequest{
				RoleUid:            "role-123",
				GroupPermissionUid: "gperm-123",
			},
			setup: func(m *servicemocks.MockRoleService) {
				m.EXPECT().AssignPermission(mock.Anything, "role-123", "gperm-123").Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(servicemocks.MockRoleService)
			if tt.setup != nil {
				tt.setup(mockSvc)
			}

			h := handler.NewRoleHandler(mockSvc, nil)
			_, err := h.AssignPermission(context.Background(), tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("AssignPermission() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
