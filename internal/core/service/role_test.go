package service

import (
	"context"
	"errors"
	"testing"

	"github.com/adityakw90/service-access/internal/core/domain/model"
	"github.com/adityakw90/service-access/internal/core/domain/param"
	"github.com/adityakw90/service-access/internal/core/port/repository"
	"github.com/adityakw90/service-access/test/mocks/repository"
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

func TestRoleService_Create(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*mocks.UnitOfWork, *mocks.RepositoryProvider, *MockUIDGenerator)
		param   param.RoleCreateParam
		want    *model.Role
		wantErr bool
		errMsg  string
	}{
		{
			name: "Happy Path",
			setup: func(m *mocks.UnitOfWork, p *mocks.RepositoryProvider, uidGen *MockUIDGenerator) {
				uidGen.MockNew = func() string { return "test-uid" }
				m.On("Do", mock.Anything, mock.AnythingOfType("func(repository.RepositoryProvider) error")).Return(nil).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(repository.RepositoryProvider) error)
					repos := &mockRepositories{role: &mockRoleRepository{}}
					// Set up the mock to return nil for Create
					repos.role.(*mockRoleRepository).On("Create", mock.Anything, mock.AnythingOfType("*model.Role")).Return(nil)

					// Call the function with our mock repositories
					fn(repos)
				})
			},
			param: param.RoleCreateParam{
				GroupID:     1,
				Name:        "admin",
				Description: "Administrator role",
			},
			want: &model.Role{
				ID:          1,
				UID:         "test-uid",
				GroupID:     1,
				Name:        "admin",
				Description: "Administrator role",
			},
			wantErr: false,
		},
		{
			name: "UnitOfWork Error",
			setup: func(m *mocks.UnitOfWork, p *mocks.RepositoryProvider, uidGen *MockUIDGenerator) {
				uidGen.MockNew = func() string { return "test-uid" }
				m.On("Do", mock.Anything, mock.AnythingOfType("func(repository.RepositoryProvider) error")).Return(errors.New("transaction error"))
			},
			param: param.RoleCreateParam{
				GroupID:     1,
				Name:        "admin",
				Description: "Administrator role",
			},
			wantErr: true,
			errMsg:  "transaction error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUoW := mocks.NewUnitOfWork(t)
			mockRepos := mocks.NewRepositoryProvider(t)
			mockPublisher := &mockPublisher{}
			mockUIDGenerator := &MockUIDGenerator{}

			tt.setup(mockUoW, mockRepos, mockUIDGenerator)

			service := NewRoleService(mockUoW, mockRepos, mockPublisher, mockUIDGenerator)
			got, err := service.Create(context.Background(), tt.param)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				if got != nil {
					assert.Equal(t, tt.want.UID, got.UID)
					assert.Equal(t, tt.want.GroupID, got.GroupID)
					assert.Equal(t, tt.want.Name, got.Name)
					assert.Equal(t, tt.want.Description, got.Description)
				}
			}

			mockUoW.AssertExpectations(t)
			mockRepos.AssertExpectations(t)
		})
	}
}
