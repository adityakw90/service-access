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

// mockPermissionRepository is a temporary mock for PermissionRepository
// This will be replaced by generated mocks once mockery is run on the repository
type mockPermissionRepository struct {
	mock.Mock
}

func (m *mockPermissionRepository) Create(ctx context.Context, permission *model.Permission) error {
	args := m.Called(ctx, permission)
	return args.Error(0)
}

func (m *mockPermissionRepository) Update(ctx context.Context, permission *model.Permission) error {
	args := m.Called(ctx, permission)
	return args.Error(0)
}

func (m *mockPermissionRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockPermissionRepository) GetByID(ctx context.Context, id int64) (*model.Permission, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Permission), args.Error(1)
}

func (m *mockPermissionRepository) GetByUID(ctx context.Context, uid string) (*model.Permission, error) {
	args := m.Called(ctx, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Permission), args.Error(1)
}

func (m *mockPermissionRepository) List(ctx context.Context, pagination *param.PaginationParam, filter *param.PermissionListFilterParam) (model.Permissions, error) {
	args := m.Called(ctx, pagination, filter)
	return args.Get(0).(model.Permissions), args.Error(1)
}

func TestPermissionService_Create(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*mocks.UnitOfWork, *mocks.RepositoryProvider, *MockUIDGenerator)
		param   param.PermissionCreateParam
		want    *model.Permission
		wantErr bool
		errMsg  string
	}{
		{
			name: "Happy Path",
			setup: func(m *mocks.UnitOfWork, p *mocks.RepositoryProvider, uidGen *MockUIDGenerator) {
				uidGen.MockNew = func() string { return "test-uid" }
				m.On("Do", mock.Anything, mock.AnythingOfType("func(repository.RepositoryProvider) error")).Return(nil).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(repository.RepositoryProvider) error)
					repos := &mockRepositories{permission: &mockPermissionRepository{}}
					// Set up the mock to return nil for Create
					repos.permission.(*mockPermissionRepository).On("Create", mock.Anything, mock.AnythingOfType("*model.Permission")).Return(nil)

					// Call the function with our mock repositories
					fn(repos)
				})
			},
			param: param.PermissionCreateParam{
				Resource:    "articles",
				Action:      "read",
				Description: "Read articles",
			},
			want: &model.Permission{
				ID:          1,
				UID:         "test-uid",
				Resource:    "articles",
				Action:      "read",
				Description: "Read articles",
			},
			wantErr: false,
		},
		{
			name: "UnitOfWork Error",
			setup: func(m *mocks.UnitOfWork, p *mocks.RepositoryProvider, uidGen *MockUIDGenerator) {
				uidGen.MockNew = func() string { return "test-uid" }
				m.On("Do", mock.Anything, mock.AnythingOfType("func(repository.RepositoryProvider) error")).Return(errors.New("transaction error"))
			},
			param: param.PermissionCreateParam{
				Resource:    "articles",
				Action:      "read",
				Description: "Read articles",
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

			service := NewPermissionService(mockUoW, mockRepos, mockPublisher, mockUIDGenerator)
			got, err := service.Create(context.Background(), tt.param)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				if got != nil {
					assert.Equal(t, tt.want.UID, got.UID)
					assert.Equal(t, tt.want.Resource, got.Resource)
					assert.Equal(t, tt.want.Action, got.Action)
					assert.Equal(t, tt.want.Description, got.Description)
				}
			}

			mockUoW.AssertExpectations(t)
			mockRepos.AssertExpectations(t)
		})
	}
}

func TestPermissionService_Get(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*mocks.RepositoryProvider)
		uid     string
		want    *model.Permission
		wantErr bool
		errMsg  string
	}{
		{
			name: "Happy Path",
			setup: func(p *mocks.RepositoryProvider) {
				mockPermRepo := &mockPermissionRepository{}
				mockPermRepo.On("GetByUID", mock.Anything, "test-uid").Return(&model.Permission{
					ID:          1,
					UID:         "test-uid",
					Resource:    "articles",
					Action:      "read",
					Description: "Read articles",
				}, nil)
				p.On("Permission").Return(mockPermRepo)
			},
			uid: "test-uid",
			want: &model.Permission{
				ID:          1,
				UID:         "test-uid",
				Resource:    "articles",
				Action:      "read",
				Description: "Read articles",
			},
			wantErr: false,
		},
		{
			name: "Not Found",
			setup: func(p *mocks.RepositoryProvider) {
				mockPermRepo := &mockPermissionRepository{}
				mockPermRepo.On("GetByUID", mock.Anything, "not-found").Return(nil, errors.New("permission not found"))
				p.On("Permission").Return(mockPermRepo)
			},
			uid:     "not-found",
			wantErr: true,
			errMsg:  "failed to get permission",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUoW := mocks.NewUnitOfWork(t)
			mockRepos := mocks.NewRepositoryProvider(t)
			mockPublisher := &mockPublisher{}
			mockUIDGenerator := &MockUIDGenerator{}

			tt.setup(mockRepos)

			service := NewPermissionService(mockUoW, mockRepos, mockPublisher, mockUIDGenerator)
			got, err := service.Get(context.Background(), tt.uid)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want.Resource, got.Resource)
			}

			mockRepos.AssertExpectations(t)
		})
	}
}

func TestPermissionService_List(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*mocks.RepositoryProvider)
		wantErr bool
	}{
		{
			name: "Happy Path",
			setup: func(p *mocks.RepositoryProvider) {
				mockPermRepo := &mockPermissionRepository{}
				mockPermRepo.On("List", mock.Anything, mock.Anything, mock.Anything).Return(model.Permissions{
					Items: []model.Permission{
						{ID: 1, UID: "test-uid-1", Resource: "articles", Action: "read"},
						{ID: 2, UID: "test-uid-2", Resource: "articles", Action: "write"},
					},
					Meta: model.Meta{Total: 2},
				}, nil)
				p.On("Permission").Return(mockPermRepo)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUoW := mocks.NewUnitOfWork(t)
			mockRepos := mocks.NewRepositoryProvider(t)
			mockPublisher := &mockPublisher{}
			mockUIDGenerator := &MockUIDGenerator{}

			tt.setup(mockRepos)

			service := NewPermissionService(mockUoW, mockRepos, mockPublisher, mockUIDGenerator)
			got, err := service.List(context.Background(), nil, nil)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, int64(2), got.Meta.Total)
			}

			mockRepos.AssertExpectations(t)
		})
	}
}
