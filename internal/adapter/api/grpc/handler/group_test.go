package handler_test

import (
	"context"
	"testing"

	"github.com/adityakw90/service-access/internal/adapter/api/grpc/handler"
	"github.com/adityakw90/service-access/internal/adapter/api/grpc/validator"
	"github.com/adityakw90/service-access/internal/core/domain/model"
	servicemocks "github.com/adityakw90/service-access/test/mocks/service"
	"github.com/adityakw90/service-access-proto/gen/go/group"
	"github.com/stretchr/testify/mock"
)

func TestGroupHandler_Create(t *testing.T) {
	tests := []struct {
		name    string
		req     *group.CreateRequest
		setup   func(*servicemocks.MockGroupService)
		want    *group.CreateResponse
		wantErr bool
	}{
		{
			name: "Happy Path",
			req: &group.CreateRequest{
				Name:        "invoice-management",
				Description: "Invoice permissions",
			},
			setup: func(m *servicemocks.MockGroupService) {
				m.EXPECT().Create(mock.Anything, mock.Anything).Return(
					&model.Group{UID: "group-123", Name: "invoice-management"}, nil,
				)
			},
			want: &group.CreateResponse{Uid: "group-123"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(servicemocks.MockGroupService)
			v := validator.New()
			if tt.setup != nil {
				tt.setup(mockSvc)
			}

			h := handler.NewGroupHandler(mockSvc, v)
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

func TestGroupHandler_List(t *testing.T) {
	tests := []struct {
		name    string
		req     *group.ListRequest
		setup   func(*servicemocks.MockGroupService)
		wantErr bool
	}{
		{
			name: "Happy Path",
			req:  &group.ListRequest{},
			setup: func(m *servicemocks.MockGroupService) {
				m.EXPECT().List(mock.Anything, mock.Anything, mock.Anything).Return(
					&model.Groups{
						Items: []model.Group{{UID: "group-1"}, {UID: "group-2"}},
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
			mockSvc := new(servicemocks.MockGroupService)
			v := validator.New()
			if tt.setup != nil {
				tt.setup(mockSvc)
			}

			h := handler.NewGroupHandler(mockSvc, v)
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

func TestGroupHandler_AssignPermission(t *testing.T) {
	tests := []struct {
		name    string
		req     *group.AssignPermissionRequest
		setup   func(*servicemocks.MockGroupService)
		wantErr bool
	}{
		{
			name: "Happy Path",
			req: &group.AssignPermissionRequest{
				GroupUid:      "group-123",
				PermissionUid: "perm-123",
			},
			setup: func(m *servicemocks.MockGroupService) {
				m.EXPECT().AssignPermission(mock.Anything, "group-123", "perm-123").Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(servicemocks.MockGroupService)
			v := validator.New()
			if tt.setup != nil {
				tt.setup(mockSvc)
			}

			h := handler.NewGroupHandler(mockSvc, v)
			_, err := h.AssignPermission(context.Background(), tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("AssignPermission() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
