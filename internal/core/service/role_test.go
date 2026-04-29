package service

import (
	"context"
	"errors"
	"testing"

	adapterobserver "github.com/adityakw90/service-access/internal/adapter/observer"
	"github.com/adityakw90/service-access/internal/core/domain/model"
	"github.com/adityakw90/service-access/internal/core/domain/param"
	"github.com/adityakw90/service-access/internal/core/domain/signal"
	"github.com/adityakw90/service-access/internal/core/port/repository"
	repomocks "github.com/adityakw90/service-access/mocks/repository"
	resolvermocks "github.com/adityakw90/service-access/mocks/resolver"
	securitymocks "github.com/adityakw90/service-access/mocks/security"
	eventmocks "github.com/adityakw90/service-access/mocks/event"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRoleService_Create(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*repomocks.MockUnitOfWork, *repomocks.MockRepositoryProvider, *securitymocks.MockUIDGenerator, *resolvermocks.MockResolverProvider, *eventmocks.MockEventPublisher)
		param   param.RoleCreateParam
		want    *model.Role
		wantErr bool
		errMsg  string
	}{
		{
			name: "Happy Path",
			setup: func(m *repomocks.MockUnitOfWork, p *repomocks.MockRepositoryProvider, uidGen *securitymocks.MockUIDGenerator, resolver *resolvermocks.MockResolverProvider, pub *eventmocks.MockEventPublisher) {
				uidGen.EXPECT().New().Return("test-uid")
				pub.EXPECT().Publish(mock.Anything, mock.Anything, mock.AnythingOfType("*event.EventRoleCreateData")).Return(nil)

				// Mock resolver to return GroupID for GroupUID
				mockGroupResolver := resolvermocks.NewMockGroupResolver(t)
				resolver.EXPECT().Group().Return(mockGroupResolver)
				mockGroupResolver.EXPECT().IDsByUIDs(mock.Anything, mock.AnythingOfType("[]string")).Return(map[string]int64{"group-uid-123": 123}, nil)

				// Mock UnitOfWork to execute the transaction
				m.EXPECT().Do(mock.Anything, mock.AnythingOfType("func(repository.RepositoryProvider) error")).RunAndReturn(func(ctx context.Context, fn func(repository.RepositoryProvider) error) error {
					mockRoleRepo := repomocks.NewMockRoleRepository(t)
					mockRoleRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*model.Role")).Return(nil).Run(func(ctx context.Context, role *model.Role) {
						// Set the ID on the role after creation
						role.ID = 1
					})
					p.EXPECT().Role().Return(mockRoleRepo)
					return fn(p)
				})
			},
			param: param.RoleCreateParam{
				GroupUID:    "group-uid-123",
				Name:        "admin",
				Description: "Administrator role",
			},
			want: &model.Role{
				ID:          1,
				UID:         "test-uid",
				GroupID:     123,
				GroupUID:    "group-uid-123",
				Name:        "admin",
				Description: "Administrator role",
			},
			wantErr: false,
		},
		{
			name: "Error - Group not found",
			setup: func(m *repomocks.MockUnitOfWork, p *repomocks.MockRepositoryProvider, uidGen *securitymocks.MockUIDGenerator, resolver *resolvermocks.MockResolverProvider, pub *eventmocks.MockEventPublisher) {
				// Note: UID generator is NOT called because we return early when group is not found

				// Mock resolver to return empty map (group not found)
				mockGroupResolver := resolvermocks.NewMockGroupResolver(t)
				resolver.EXPECT().Group().Return(mockGroupResolver)
				mockGroupResolver.EXPECT().IDsByUIDs(mock.Anything, mock.AnythingOfType("[]string")).Return(map[string]int64{}, nil)
			},
			param: param.RoleCreateParam{
				GroupUID:    "non-existent-group",
				Name:        "admin",
				Description: "Administrator role",
			},
			wantErr: true,
			errMsg:  "group not found",
		},
		{
			name: "Error - UnitOfWork transaction error",
			setup: func(m *repomocks.MockUnitOfWork, p *repomocks.MockRepositoryProvider, uidGen *securitymocks.MockUIDGenerator, resolver *resolvermocks.MockResolverProvider, pub *eventmocks.MockEventPublisher) {
				// Note: UID generator is NOT called because UoW.Do() returns error before executing the callback

				// Mock resolver (will be called before the error)
				mockGroupResolver := resolvermocks.NewMockGroupResolver(t)
				resolver.EXPECT().Group().Return(mockGroupResolver)
				mockGroupResolver.EXPECT().IDsByUIDs(mock.Anything, mock.AnythingOfType("[]string")).Return(map[string]int64{"group-uid-123": 123}, nil)

				// Mock UnitOfWork to return error without executing the callback
				m.EXPECT().Do(mock.Anything, mock.AnythingOfType("func(repository.RepositoryProvider) error")).Return(errors.New("transaction error"))
			},
			param: param.RoleCreateParam{
				GroupUID:    "group-uid-123",
				Name:        "admin",
				Description: "Administrator role",
			},
			wantErr: true,
			errMsg:  "failed to create role",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUoW := repomocks.NewMockUnitOfWork(t)
			mockRepos := repomocks.NewMockRepositoryProvider(t)
			mockPublisher := eventmocks.NewMockEventPublisher(t)
			mockUIDGenerator := securitymocks.NewMockUIDGenerator(t)
			mockResolverProvider := resolvermocks.NewMockResolverProvider(t)
			mockObserver := adapterobserver.NewNoopObserver[signal.SignalRole]()

			tt.setup(mockUoW, mockRepos, mockUIDGenerator, mockResolverProvider, mockPublisher)

			service := NewRoleService(mockUoW, mockRepos, mockPublisher, mockUIDGenerator, mockResolverProvider, nil, mockObserver, adapterobserver.NewNoopObserver[signal.SignalRolePermission]())
			got, err := service.Create(context.Background(), tt.param)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				if got != nil {
					assert.Equal(t, tt.want.UID, got.UID)
					assert.Equal(t, tt.want.GroupID, got.GroupID)
					assert.Equal(t, tt.want.GroupUID, got.GroupUID)
					assert.Equal(t, tt.want.Name, got.Name)
					assert.Equal(t, tt.want.Description, got.Description)
				}
			}

			mockUoW.AssertExpectations(t)
			mockRepos.AssertExpectations(t)
		})
	}
}
