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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockGroupRepository is a temporary mock for GroupRepository
// This will be replaced by generated mocks once mockery is run on the repository
type mockGroupRepository struct {
	mock.Mock
}

func (m *mockGroupRepository) Create(ctx context.Context, group *model.Group) error {
	args := m.Called(ctx, group)
	return args.Error(0)
}

func (m *mockGroupRepository) Update(ctx context.Context, group *model.Group) error {
	args := m.Called(ctx, group)
	return args.Error(0)
}

func (m *mockGroupRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockGroupRepository) GetByID(ctx context.Context, id int64) (*model.Group, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Group), args.Error(1)
}

func (m *mockGroupRepository) GetByUID(ctx context.Context, uid string) (*model.Group, error) {
	args := m.Called(ctx, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Group), args.Error(1)
}

func (m *mockGroupRepository) List(ctx context.Context, pagination *param.PaginationParam, filter *param.GroupListFilterParam) (model.Groups, error) {
	args := m.Called(ctx, pagination, filter)
	return args.Get(0).(model.Groups), args.Error(1)
}

func (m *mockGroupRepository) ListPermission(ctx context.Context, groupID int64, pagination *param.PaginationParam, filter *param.GroupPermissionListFilterParam) (model.GroupPermissions, error) {
	args := m.Called(ctx, groupID, pagination, filter)
	return args.Get(0).(model.GroupPermissions), args.Error(1)
}

func (m *mockGroupRepository) AddPermission(ctx context.Context, groupID int64, permissionID int64, uid string) error {
	args := m.Called(ctx, groupID, permissionID, uid)
	return args.Error(0)
}

func (m *mockGroupRepository) RemovePermission(ctx context.Context, groupID int64, permissionID int64) error {
	args := m.Called(ctx, groupID, permissionID)
	return args.Error(0)
}

func (m *mockGroupRepository) ReplacePermission(ctx context.Context, groupID int64, permissionIDs []int64, uids []string) error {
	args := m.Called(ctx, groupID, permissionIDs, uids)
	return args.Error(0)
}

func (m *mockGroupRepository) GetPermissionByID(ctx context.Context, groupPermissionID int64) (*model.GroupPermission, error) {
	args := m.Called(ctx, groupPermissionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.GroupPermission), args.Error(1)
}

func (m *mockGroupRepository) GetPermissionByGroupIDAndPermissionUID(ctx context.Context, groupID int64, permissionUID string) (*model.GroupPermission, error) {
	args := m.Called(ctx, groupID, permissionUID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.GroupPermission), args.Error(1)
}

func TestGroupService_Create(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*repomocks.MockUnitOfWork, *repomocks.MockRepositoryProvider, *securitymocks.MockUIDGenerator)
		param   param.GroupCreateParam
		want    *model.Group
		wantErr bool
		errMsg  string
	}{
		{
			name: "Happy Path",
			setup: func(m *repomocks.MockUnitOfWork, p *repomocks.MockRepositoryProvider, uidGen *securitymocks.MockUIDGenerator) {
				uidGen.On("New").Return("test-uid")
				m.On("Do", mock.Anything, mock.AnythingOfType("func(repository.RepositoryProvider) error")).Return(nil).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(repository.RepositoryProvider) error)
					repos := &mockRepositories{group: &mockGroupRepository{}}
					// Set up the mock to return nil for Create
					repos.group.(*mockGroupRepository).On("Create", mock.Anything, mock.AnythingOfType("*model.Group")).Return(nil)

					// Call the function with our mock repositories
					fn(repos)
				})
			},
			param: param.GroupCreateParam{
				Name:        "admin",
				Description: "Administrators",
			},
			want: &model.Group{
				ID:          1,
				UID:         "test-uid",
				Name:        "admin",
				Description: "Administrators",
			},
			wantErr: false,
		},
		{
			name: "UnitOfWork Error",
			setup: func(m *repomocks.MockUnitOfWork, p *repomocks.MockRepositoryProvider, uidGen *securitymocks.MockUIDGenerator) {
				uidGen.On("New").Return("test-uid")
				m.On("Do", mock.Anything, mock.AnythingOfType("func(repository.RepositoryProvider) error")).Return(errors.New("transaction error"))
			},
			param: param.GroupCreateParam{
				Name:        "admin",
				Description: "Administrators",
			},
			wantErr: true,
			errMsg:  "transaction error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUoW := repomocks.NewMockUnitOfWork(t)
			mockRepos := repomocks.NewMockRepositoryProvider(t)
			mockPublisher := &mockPublisher{}
			mockUIDGenerator := securitymocks.NewMockUIDGenerator(t)
			mockResolverProvider := resolvermocks.NewMockResolverProvider(t)
			mockObserver := adapterobserver.NewNoopObserver[signal.SignalGroup]()

			tt.setup(mockUoW, mockRepos, mockUIDGenerator)

			service := NewGroupService(mockUoW, mockRepos, mockPublisher, mockUIDGenerator, mockResolverProvider, mockObserver, adapterobserver.NewNoopObserver[signal.SignalGroupPermission]())
			got, err := service.Create(context.Background(), tt.param)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				if got != nil {
					assert.Equal(t, tt.want.UID, got.UID)
					assert.Equal(t, tt.want.Name, got.Name)
					assert.Equal(t, tt.want.Description, got.Description)
				}
			}

			mockUoW.AssertExpectations(t)
			mockRepos.AssertExpectations(t)
		})
	}
}

func TestGroupService_Get(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*repomocks.MockRepositoryProvider)
		uid     string
		want    *model.Group
		wantErr bool
		errMsg  string
	}{
		{
			name: "Happy Path",
			setup: func(p *repomocks.MockRepositoryProvider) {
				mockGroupRepo := &mockGroupRepository{}
				mockGroupRepo.On("GetByID", mock.Anything, int64(1)).Return(&model.Group{
					ID:          1,
					UID:         "test-uid",
					Name:        "admin",
					Description: "Administrators",
				}, nil)
				p.On("Group").Return(mockGroupRepo)
			},
			uid: "test-uid",
			want: &model.Group{
				ID:          1,
				UID:         "test-uid",
				Name:        "admin",
				Description: "Administrators",
			},
			wantErr: false,
		},
		{
			name: "Not Found",
			setup: func(p *repomocks.MockRepositoryProvider) {
				mockGroupRepo := &mockGroupRepository{}
				mockGroupRepo.On("GetByID", mock.Anything, int64(1)).Return(nil, errors.New("group not found"))
				p.On("Group").Return(mockGroupRepo)
			},
			uid:     "not-found",
			wantErr: true,
			errMsg:  "failed to get group",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUoW := repomocks.NewMockUnitOfWork(t)
			mockRepos := repomocks.NewMockRepositoryProvider(t)
			mockPublisher := &mockPublisher{}
			mockUIDGenerator := securitymocks.NewMockUIDGenerator(t)
			mockResolverProvider := resolvermocks.NewMockResolverProvider(t)
			mockObserver := adapterobserver.NewNoopObserver[signal.SignalGroup]()

			// Set up resolver mock
			mockGroupResolver := resolvermocks.NewMockGroupResolver(t)
			mockGroupResolver.On("IDsByUIDs", mock.Anything, []string{tt.uid}).Return(map[string]int64{tt.uid: 1}, nil)
			mockResolverProvider.On("Group").Return(mockGroupResolver)

			tt.setup(mockRepos)

			service := NewGroupService(mockUoW, mockRepos, mockPublisher, mockUIDGenerator, mockResolverProvider, mockObserver, adapterobserver.NewNoopObserver[signal.SignalGroupPermission]())
			got, err := service.Get(context.Background(), tt.uid)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want.Name, got.Name)
			}

			mockRepos.AssertExpectations(t)
		})
	}
}

func TestGroupService_List(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*repomocks.MockRepositoryProvider)
		wantErr bool
	}{
		{
			name: "Happy Path",
			setup: func(p *repomocks.MockRepositoryProvider) {
				mockGroupRepo := &mockGroupRepository{}
				mockGroupRepo.On("List", mock.Anything, mock.Anything, mock.Anything).Return(model.Groups{
					Items: []model.Group{
						{ID: 1, UID: "test-uid-1", Name: "admin"},
						{ID: 2, UID: "test-uid-2", Name: "user"},
					},
					Meta: model.Meta{Total: 2},
				}, nil)
				p.On("Group").Return(mockGroupRepo)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUoW := repomocks.NewMockUnitOfWork(t)
			mockRepos := repomocks.NewMockRepositoryProvider(t)
			mockPublisher := &mockPublisher{}
			mockUIDGenerator := securitymocks.NewMockUIDGenerator(t)
			mockResolverProvider := resolvermocks.NewMockResolverProvider(t)
			mockObserver := adapterobserver.NewNoopObserver[signal.SignalGroup]()

			tt.setup(mockRepos)

			service := NewGroupService(mockUoW, mockRepos, mockPublisher, mockUIDGenerator, mockResolverProvider, mockObserver, adapterobserver.NewNoopObserver[signal.SignalGroupPermission]())
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
