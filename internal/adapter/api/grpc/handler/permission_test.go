package handler_test

import (
	"context"
	"testing"

	"github.com/adityakw90/service-access/internal/adapter/api/grpc/handler"
	"github.com/adityakw90/service-access/internal/adapter/api/grpc/validator"
	"github.com/adityakw90/service-access/internal/core/domain/model"
	domainErrors "github.com/adityakw90/service-access/internal/core/domain/errors"
	servicemocks "github.com/adityakw90/service-access/test/mocks/service"
	"github.com/adityakw90/service-access-proto/gen/go/permission"
	"github.com/stretchr/testify/mock"
)

func TestPermissionHandler_Create(t *testing.T) {
	tests := []struct {
		name       string
		req        *permission.CreateRequest
		setup      func(*servicemocks.MockPermissionService)
		want       *permission.CreateResponse
		wantErr    bool
		wantErrCode string
	}{
		{
			name: "Happy Path",
			req: &permission.CreateRequest{
				Resource:    "invoices",
				Action:      "read",
				Description: "Read invoices",
			},
			setup: func(m *servicemocks.MockPermissionService) {
				m.EXPECT().Create(mock.Anything, mock.AnythingOfType("param.PermissionCreateParam")).Return(
					&model.Permission{UID: "perm-123"}, nil,
				)
			},
			want: &permission.CreateResponse{Uid: "perm-123"},
		},
		{
			name: "Service Error - Not Found",
			req: &permission.CreateRequest{
				Resource: "invoices",
				Action:   "read",
			},
			setup: func(m *servicemocks.MockPermissionService) {
				m.EXPECT().Create(mock.Anything, mock.Anything).Return(
					nil, domainErrors.ErrNotFound,
				)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(servicemocks.MockPermissionService)
			if tt.setup != nil {
				tt.setup(mockSvc)
			}

			h := handler.NewPermissionHandler(mockSvc, validator.New())
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

func TestPermissionHandler_Get(t *testing.T) {
	tests := []struct {
		name    string
		req     *permission.GetRequest
		setup   func(*servicemocks.MockPermissionService)
		want    *permission.Permission
		wantErr bool
	}{
		{
			name: "Happy Path",
			req:  &permission.GetRequest{Uid: "perm-123"},
			setup: func(m *servicemocks.MockPermissionService) {
				m.EXPECT().Get(mock.Anything, "perm-123").Return(
					&model.Permission{UID: "perm-123", Resource: "invoices", Action: "read"}, nil,
				)
			},
			want: &permission.Permission{Uid: "perm-123", Resource: "invoices", Action: "read"},
		},
		{
			name:    "Not Found",
			req:     &permission.GetRequest{Uid: "unknown"},
			setup: func(m *servicemocks.MockPermissionService) {
				m.EXPECT().Get(mock.Anything, "unknown").Return(nil, domainErrors.ErrNotFound)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(servicemocks.MockPermissionService)
			if tt.setup != nil {
				tt.setup(mockSvc)
			}

			h := handler.NewPermissionHandler(mockSvc, validator.New())
			got, err := h.Get(context.Background(), tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got.Uid != tt.want.Uid {
				t.Errorf("Get() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPermissionHandler_List(t *testing.T) {
	tests := []struct {
		name    string
		req     *permission.ListRequest
		setup   func(*servicemocks.MockPermissionService)
		wantErr bool
	}{
		{
			name: "Happy Path",
			req:  &permission.ListRequest{},
			setup: func(m *servicemocks.MockPermissionService) {
				m.EXPECT().List(mock.Anything, mock.Anything, mock.Anything).Return(
					&model.Permissions{
						Items: []model.Permission{{UID: "perm-1"}, {UID: "perm-2"}},
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
			mockSvc := new(servicemocks.MockPermissionService)
			if tt.setup != nil {
				tt.setup(mockSvc)
			}

			h := handler.NewPermissionHandler(mockSvc, validator.New())
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

func TestPermissionHandler_Update(t *testing.T) {
	tests := []struct {
		name    string
		req     *permission.UpdateRequest
		setup   func(*servicemocks.MockPermissionService)
		wantErr bool
	}{
		{
			name: "Happy Path",
			req:  &permission.UpdateRequest{Uid: "perm-123", Resource: "invoices", Action: "write"},
			setup: func(m *servicemocks.MockPermissionService) {
				m.EXPECT().Update(mock.Anything, "perm-123", mock.Anything).Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(servicemocks.MockPermissionService)
			if tt.setup != nil {
				tt.setup(mockSvc)
			}

			h := handler.NewPermissionHandler(mockSvc, validator.New())
			_, err := h.Update(context.Background(), tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPermissionHandler_Delete(t *testing.T) {
	tests := []struct {
		name    string
		req     *permission.DeleteRequest
		setup   func(*servicemocks.MockPermissionService)
		wantErr bool
	}{
		{
			name: "Happy Path",
			req:  &permission.DeleteRequest{Uid: "perm-123"},
			setup: func(m *servicemocks.MockPermissionService) {
				m.EXPECT().Delete(mock.Anything, "perm-123").Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(servicemocks.MockPermissionService)
			if tt.setup != nil {
				tt.setup(mockSvc)
			}

			h := handler.NewPermissionHandler(mockSvc, validator.New())
			_, err := h.Delete(context.Background(), tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
