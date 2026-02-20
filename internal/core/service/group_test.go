package service

import (
	"context"
	"errors"
	"testing"

	"github.com/adityakw90/service-access/internal/core/domain/event"
	"github.com/adityakw90/service-access/internal/core/domain/model"
	"github.com/adityakw90/service-access/internal/core/domain/param"
	"github.com/adityakw90/service-access/internal/core/port/repository"
	"github.com/adityakw90/service-access/mocks"
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

func (m *mockGroupRepository) AddPermission(ctx context.Context, groupID int64, permissionID int64) error {
	args := m.Called(ctx, groupID, permissionID)
	return args.Error(0)
}

func (m *mockGroupRepository) RemovePermission(ctx context.Context, groupID int64, permissionID int64) error {
	args := m.Called(ctx, groupID, permissionID)
	return args.Error(0)
}

func (m *mockGroupRepository) ReplacePermission(ctx context.Context, groupID int64, permissionIDs []int64) error {
	args := m.Called(ctx, groupID, permissionIDs)
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

// mockRepositories implements repository.RepositoryProvider for testing
type mockRepositories struct {
	group      repository.GroupRepository
	permission repository.PermissionRepository
	role       repository.RoleRepository
	subject    repository.SubjectRepository
}

func (m *mockRepositories) Group() repository.GroupRepository {
	return m.group
}

func (m *mockRepositories) Permission() repository.PermissionRepository {
	return m.permission
}

func (m *mockRepositories) Role() repository.RoleRepository {
	return m.role
}

func (m *mockRepositories) Subject() repository.SubjectRepository {
	return m.subject
}

// mockPublisher is a simple mock for event.Publisher
type mockPublisher struct {
	publishedEvents []publishedEvent
}

type publishedEvent struct {
	eventType event.EventType
	eventData any
}

func (m *mockPublisher) Publish(ctx context.Context, eventType event.EventType, eventData any) error {
	m.publishedEvents = append(m.publishedEvents, publishedEvent{eventType: eventType, eventData: eventData})
	return nil
}

func (m *mockPublisher) Close() error {
	return nil
}

func TestGroupService_Create(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*mocks.UnitOfWork, *mocks.RepositoryProvider, *MockUIDGenerator)
		param   param.GroupCreateParam
		want    *model.Group
		wantErr bool
		errMsg  string
	}{
		{
			name: "Happy Path",
			setup: func(m *mocks.UnitOfWork, p *mocks.RepositoryProvider, uidGen *MockUIDGenerator) {
				uidGen.MockNew = func() string { return "test-uid" }
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
			setup: func(m *mocks.UnitOfWork, p *mocks.RepositoryProvider, uidGen *MockUIDGenerator) {
				uidGen.MockNew = func() string { return "test-uid" }
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
			mockUoW := mocks.NewUnitOfWork(t)
			mockRepos := mocks.NewRepositoryProvider(t)
			mockPublisher := &mockPublisher{}
			mockUIDGenerator := &MockUIDGenerator{}

			tt.setup(mockUoW, mockRepos, mockUIDGenerator)

			service := NewGroupService(mockUoW, mockRepos, mockPublisher, mockUIDGenerator)
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
		setup   func(*mocks.RepositoryProvider)
		uid     string
		want    *model.Group
		wantErr bool
		errMsg  string
	}{
		{
			name: "Happy Path",
			setup: func(p *mocks.RepositoryProvider) {
				mockGroupRepo := &mockGroupRepository{}
				mockGroupRepo.On("GetByUID", mock.Anything, "test-uid").Return(&model.Group{
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
			setup: func(p *mocks.RepositoryProvider) {
				mockGroupRepo := &mockGroupRepository{}
				mockGroupRepo.On("GetByUID", mock.Anything, "not-found").Return(nil, errors.New("group not found"))
				p.On("Group").Return(mockGroupRepo)
			},
			uid:     "not-found",
			wantErr: true,
			errMsg:  "failed to get group",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUoW := mocks.NewUnitOfWork(t)
			mockRepos := mocks.NewRepositoryProvider(t)
			mockPublisher := &mockPublisher{}
			mockUIDGenerator := &MockUIDGenerator{}

			tt.setup(mockRepos)

			service := NewGroupService(mockUoW, mockRepos, mockPublisher, mockUIDGenerator)
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
		setup   func(*mocks.RepositoryProvider)
		wantErr bool
	}{
		{
			name: "Happy Path",
			setup: func(p *mocks.RepositoryProvider) {
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
			mockUoW := mocks.NewUnitOfWork(t)
			mockRepos := mocks.NewRepositoryProvider(t)
			mockPublisher := &mockPublisher{}
			mockUIDGenerator := &MockUIDGenerator{}

			tt.setup(mockRepos)

			service := NewGroupService(mockUoW, mockRepos, mockPublisher, mockUIDGenerator)
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
