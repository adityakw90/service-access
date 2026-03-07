package service

import (
	"context"
	"errors"
	"testing"

	"github.com/adityakw90/service-access/internal/core/domain/model"
	"github.com/adityakw90/service-access/internal/core/domain/param"
	"github.com/adityakw90/service-access/internal/core/domain/signal"
	adapterobserver "github.com/adityakw90/service-access/internal/adapter/observer"
	"github.com/adityakw90/service-access/internal/core/port/repository"
	repomocks "github.com/adityakw90/service-access/test/mocks/repository"
	resolvermocks "github.com/adityakw90/service-access/test/mocks/resolver"
	securitymocks "github.com/adityakw90/service-access/test/mocks/security"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockRoleRepository is a temporary mock for RoleRepository
// This will be replaced by generated mocks once mockery is run on the repository
type mockRoleRepository struct {
	mock.Mock
}

func (m *mockRoleRepository) Create(ctx context.Context, role *model.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *mockRoleRepository) Update(ctx context.Context, role *model.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *mockRoleRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockRoleRepository) GetByID(ctx context.Context, id int64) (*model.Role, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Role), args.Error(1)
}

func (m *mockRoleRepository) GetByUID(ctx context.Context, uid string) (*model.Role, error) {
	args := m.Called(ctx, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Role), args.Error(1)
}

func (m *mockRoleRepository) List(ctx context.Context, pagination *param.PaginationParam, filter *param.RoleListFilterParam) (model.Roles, error) {
	args := m.Called(ctx, pagination, filter)
	return args.Get(0).(model.Roles), args.Error(1)
}

func (m *mockRoleRepository) ListPermission(ctx context.Context, roleID int64, pagination *param.PaginationParam, filter *param.RolePermissionListFilterParam) (model.RolePermissions, error) {
	args := m.Called(ctx, roleID, pagination, filter)
	return args.Get(0).(model.RolePermissions), args.Error(1)
}

func (m *mockRoleRepository) AddPermission(ctx context.Context, roleID int64, groupPermissionID int64) error {
	args := m.Called(ctx, roleID, groupPermissionID)
	return args.Error(0)
}

func (m *mockRoleRepository) RemovePermission(ctx context.Context, roleID int64, groupPermissionID int64) error {
	args := m.Called(ctx, roleID, groupPermissionID)
	return args.Error(0)
}

func (m *mockRoleRepository) ReplacePermission(ctx context.Context, roleID int64, groupPermissionIDs []int64) error {
	args := m.Called(ctx, roleID, groupPermissionIDs)
	return args.Error(0)
}

func (m *mockRoleRepository) GetAllPermissions(ctx context.Context, roleID int64) ([]model.Permission, error) {
	args := m.Called(ctx, roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Permission), args.Error(1)
}

func TestRoleService_Create(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*repomocks.MockUnitOfWork, *repomocks.MockRepositoryProvider, *securitymocks.MockUIDGenerator, *resolvermocks.MockResolverProvider)
		param   param.RoleCreateParam
		want    *model.Role
		wantErr bool
		errMsg  string
	}{
		{
			name: "Happy Path",
			setup: func(m *repomocks.MockUnitOfWork, p *repomocks.MockRepositoryProvider, uidGen *securitymocks.MockUIDGenerator, resolver *resolvermocks.MockResolverProvider) {
				uidGen.On("New").Return("test-uid")

				// Mock resolver to return GroupID for GroupUID
				mockGroupResolver := resolvermocks.NewMockGroupResolver(t)
				resolver.EXPECT().Group().Return(mockGroupResolver)
				mockGroupResolver.On("IDsByUIDs", mock.Anything, mock.AnythingOfType("[]string")).Return(map[string]int64{"group-uid-123": 123}, nil)

				// Mock UnitOfWork to execute the transaction
				m.EXPECT().Do(mock.Anything, mock.AnythingOfType("func(repository.RepositoryProvider) error")).RunAndReturn(func(ctx context.Context, fn func(repository.RepositoryProvider) error) error {
					repos := &mockRepositories{role: &mockRoleRepository{}}
					// Set up the mock to return nil for Create
					repos.role.(*mockRoleRepository).On("Create", mock.Anything, mock.AnythingOfType("*model.Role")).Return(nil).Run(func(args mock.Arguments) {
						// Set the ID on the role after creation
						role := args.Get(1).(*model.Role)
						role.ID = 1
					})

					// Call the function with our mock repositories
					return fn(repos)
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
			setup: func(m *repomocks.MockUnitOfWork, p *repomocks.MockRepositoryProvider, uidGen *securitymocks.MockUIDGenerator, resolver *resolvermocks.MockResolverProvider) {
				// Note: UID generator is NOT called because we return early when group is not found

				// Mock resolver to return empty map (group not found)
				mockGroupResolver := resolvermocks.NewMockGroupResolver(t)
				resolver.EXPECT().Group().Return(mockGroupResolver)
				mockGroupResolver.On("IDsByUIDs", mock.Anything, mock.AnythingOfType("[]string")).Return(map[string]int64{}, nil)
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
			setup: func(m *repomocks.MockUnitOfWork, p *repomocks.MockRepositoryProvider, uidGen *securitymocks.MockUIDGenerator, resolver *resolvermocks.MockResolverProvider) {
				// Note: UID generator is NOT called because UoW.Do() returns error before executing the callback

				// Mock resolver (will be called before the error)
				mockGroupResolver := resolvermocks.NewMockGroupResolver(t)
				resolver.EXPECT().Group().Return(mockGroupResolver)
				mockGroupResolver.On("IDsByUIDs", mock.Anything, mock.AnythingOfType("[]string")).Return(map[string]int64{"group-uid-123": 123}, nil)

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
			mockPublisher := &mockPublisher{}
			mockUIDGenerator := securitymocks.NewMockUIDGenerator(t)
			mockResolverProvider := resolvermocks.NewMockResolverProvider(t)
			mockObserver := adapterobserver.NewNoopObserver[signal.SignalRole]()

			tt.setup(mockUoW, mockRepos, mockUIDGenerator, mockResolverProvider)

			service := NewRoleService(mockUoW, mockRepos, mockPublisher, mockUIDGenerator, mockResolverProvider, mockObserver, adapterobserver.NewNoopObserver[signal.SignalRolePermission]())
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
