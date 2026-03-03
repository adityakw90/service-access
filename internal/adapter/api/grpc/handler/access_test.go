package handler_test

import (
	"context"
	"testing"

	"github.com/adityakw90/service-access/internal/adapter/api/grpc/handler"
	"github.com/adityakw90/service-access/internal/adapter/api/grpc/validator"
	accessmocks "github.com/adityakw90/service-access/test/mocks/service"
	"github.com/adityakw90/service-access-proto/gen/go/access"
	"github.com/stretchr/testify/mock"
)

func TestAccessHandler_CheckAccess(t *testing.T) {
	tests := []struct {
		name        string
		req         *access.CheckAccessRequest
		setup       func(*accessmocks.MockAccessService)
		wantAllowed bool
		wantErr     bool
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
			v := validator.New()
			if tt.setup != nil {
				tt.setup(mockSvc)
			}

			h := handler.NewAccessHandler(mockSvc, v)
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
