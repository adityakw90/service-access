package service

import (
	"context"
	"testing"

	"github.com/adityakw90/service-access/internal/core/domain/event"
	"github.com/adityakw90/service-access/internal/core/domain/model"
	"github.com/adityakw90/service-access/internal/core/domain/param"
	"github.com/adityakw90/service-access/internal/core/domain/signal"
	"github.com/adityakw90/service-access/internal/core/port/repository"

	eventmocks "github.com/adityakw90/service-access/mocks/event"
	repomocks "github.com/adityakw90/service-access/mocks/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// mockSubjectObserver is a simple mock for the ServiceObserver[signal.SignalSubject] interface
type mockSubjectObserver struct {
	mock.Mock
}

func (m *mockSubjectObserver) OnSignal(ctx context.Context, sig signal.SignalType, data signal.SignalSubject, err error) {
	m.Called(ctx, sig, data, err)
}

func TestSubjectService_GetRoles(t *testing.T) {
	tests := []struct {
		name        string
		subjectID   string
		subjectType string
		setup       func(*repomocks.MockSubjectRepository)
		want        []model.Role
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "Success - Returns roles",
			subjectID:   "user-123",
			subjectType: "user",
			setup: func(m *repomocks.MockSubjectRepository) {
				m.EXPECT().GetAllRoles(mock.Anything, "user-123", "user").Return([]model.Role{
					{ID: 1, UID: "role-1", Name: "Admin", GroupID: 1, GroupUID: "group-1"},
				}, nil)
			},
			want: []model.Role{
				{ID: 1, UID: "role-1", Name: "Admin", GroupID: 1, GroupUID: "group-1"},
			},
			wantErr: false,
		},
		{
			name:        "Success - Empty roles",
			subjectID:   "user-456",
			subjectType: "user",
			setup: func(m *repomocks.MockSubjectRepository) {
				m.EXPECT().GetAllRoles(mock.Anything, "user-456", "user").Return([]model.Role{}, nil)
			},
			want:    []model.Role{},
			wantErr: false,
		},
		{
			name:        "Error - Repository failure",
			subjectID:   "user-789",
			subjectType: "user",
			setup: func(m *repomocks.MockSubjectRepository) {
				m.EXPECT().GetAllRoles(mock.Anything, "user-789", "user").Return(nil, assert.AnError)
			},
			wantErr: true,
			errMsg:  "failed to get roles",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			mockRepo := repomocks.NewMockSubjectRepository(t)
			mockObserver := new(mockSubjectObserver)
			mockObserver.On("OnSignal", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

			tt.setup(mockRepo)

			mockObserver.On("OnSignal", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

			mockProvider := new(repomocks.MockRepositoryProvider)
			mockProvider.EXPECT().Subject().Return(mockRepo)

			svc := NewSubjectService(nil, mockProvider, nil, nil, mockObserver)

			got, err := svc.GetRoles(ctx, tt.subjectID, tt.subjectType)

			if tt.wantErr {
				require.Error(t, err)
				// Error occurred as expected
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSubjectService_GetGroups(t *testing.T) {
	tests := []struct {
		name        string
		subjectID   string
		subjectType string
		setup       func(*repomocks.MockSubjectRepository)
		want        []model.Group
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "Success - Returns groups",
			subjectID:   "user-123",
			subjectType: "user",
			setup: func(m *repomocks.MockSubjectRepository) {
				m.EXPECT().GetAllGroups(mock.Anything, "user-123", "user").Return([]model.Group{
					{ID: 1, UID: "group-1", Name: "Admins"},
				}, nil)
			},
			want: []model.Group{
				{ID: 1, UID: "group-1", Name: "Admins"},
			},
			wantErr: false,
		},
		{
			name:        "Success - Empty groups",
			subjectID:   "user-456",
			subjectType: "service",
			setup: func(m *repomocks.MockSubjectRepository) {
				m.EXPECT().GetAllGroups(mock.Anything, "user-456", "service").Return([]model.Group{}, nil)
			},
			want:    []model.Group{},
			wantErr: false,
		},
		{
			name:        "Error - Repository failure",
			subjectID:   "user-789",
			subjectType: "user",
			setup: func(m *repomocks.MockSubjectRepository) {
				m.EXPECT().GetAllGroups(mock.Anything, "user-789", "user").Return(nil, assert.AnError)
			},
			wantErr: true,
			errMsg:  "failed to get groups",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			mockRepo := repomocks.NewMockSubjectRepository(t)
			mockObserver := new(mockSubjectObserver)
			mockObserver.On("OnSignal", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

			tt.setup(mockRepo)
			mockObserver.On("OnSignal", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

			mockProvider := new(repomocks.MockRepositoryProvider)
			mockProvider.EXPECT().Subject().Return(mockRepo)

			svc := NewSubjectService(nil, mockProvider, nil, nil, mockObserver)

			got, err := svc.GetGroups(ctx, tt.subjectID, tt.subjectType)

			if tt.wantErr {
				require.Error(t, err)
				// Error occurred as expected
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSubjectService_GetPermissions(t *testing.T) {
	tests := []struct {
		name        string
		subjectID   string
		subjectType string
		setup       func(*repomocks.MockSubjectRepository)
		want        []model.Permission
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "Success - Returns permissions",
			subjectID:   "user-123",
			subjectType: "user",
			setup: func(m *repomocks.MockSubjectRepository) {
				m.EXPECT().GetAllPermissions(mock.Anything, "user-123", "user").Return([]model.Permission{
					{ID: 1, UID: "perm-1", Resource: "resource", Action: "read"},
				}, nil)
			},
			want: []model.Permission{
				{ID: 1, UID: "perm-1", Resource: "resource", Action: "read"},
			},
			wantErr: false,
		},
		{
			name:        "Success - Empty permissions",
			subjectID:   "user-456",
			subjectType: "service",
			setup: func(m *repomocks.MockSubjectRepository) {
				m.EXPECT().GetAllPermissions(mock.Anything, "user-456", "service").Return([]model.Permission{}, nil)
			},
			want:    []model.Permission{},
			wantErr: false,
		},
		{
			name:        "Error - Repository failure",
			subjectID:   "user-789",
			subjectType: "user",
			setup: func(m *repomocks.MockSubjectRepository) {
				m.EXPECT().GetAllPermissions(mock.Anything, "user-789", "user").Return(nil, assert.AnError)
			},
			wantErr: true,
			errMsg:  "failed to get permissions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			mockRepo := repomocks.NewMockSubjectRepository(t)
			mockObserver := new(mockSubjectObserver)
			mockObserver.On("OnSignal", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

			tt.setup(mockRepo)
			mockObserver.On("OnSignal", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

			mockProvider := new(repomocks.MockRepositoryProvider)
			mockProvider.EXPECT().Subject().Return(mockRepo)

			svc := NewSubjectService(nil, mockProvider, nil, nil, mockObserver)

			got, err := svc.GetPermissions(ctx, tt.subjectID, tt.subjectType)

			if tt.wantErr {
				require.Error(t, err)
				// Error occurred as expected
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSubjectService_GetFullProfile(t *testing.T) {
	tests := []struct {
		name        string
		subjectID   string
		subjectType string
		setup       func(*repomocks.MockSubjectRepository)
		want        *model.SubjectProfile
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "Success - Returns full profile",
			subjectID:   "user-123",
			subjectType: "user",
			setup: func(m *repomocks.MockSubjectRepository) {
				m.EXPECT().GetAllGroups(mock.Anything, "user-123", "user").Return([]model.Group{
					{ID: 1, UID: "group-1", Name: "Admins"},
				}, nil)
				m.EXPECT().GetAllRoles(mock.Anything, "user-123", "user").Return([]model.Role{
					{ID: 1, UID: "role-1", Name: "Admin", GroupID: 1, GroupUID: "group-1"},
				}, nil)
				m.EXPECT().GetAllPermissions(mock.Anything, "user-123", "user").Return([]model.Permission{
					{ID: 1, UID: "perm-1", Resource: "resource", Action: "read"},
				}, nil)
			},
			want: &model.SubjectProfile{
				Groups: []model.Group{
					{ID: 1, UID: "group-1", Name: "Admins"},
				},
				Roles: []model.Role{
					{ID: 1, UID: "role-1", Name: "Admin", GroupID: 1, GroupUID: "group-1"},
				},
				Permissions: []model.Permission{
					{ID: 1, UID: "perm-1", Resource: "resource", Action: "read"},
				},
			},
			wantErr: false,
		},
		{
			name:        "Error - Groups repository failure",
			subjectID:   "user-456",
			subjectType: "user",
			setup: func(m *repomocks.MockSubjectRepository) {
				m.EXPECT().GetAllGroups(mock.Anything, "user-456", "user").Return(nil, assert.AnError)
				m.EXPECT().GetAllRoles(mock.Anything, "user-456", "user").Return([]model.Role{}, nil)
				m.EXPECT().GetAllPermissions(mock.Anything, "user-456", "user").Return([]model.Permission{}, nil)
			},
			wantErr: true,
			errMsg:  "failed to get groups",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			mockRepo := repomocks.NewMockSubjectRepository(t)
			mockObserver := new(mockSubjectObserver)
			mockObserver.On("OnSignal", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

			tt.setup(mockRepo)
			mockObserver.On("OnSignal", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

			mockProvider := new(repomocks.MockRepositoryProvider)
			mockProvider.EXPECT().Subject().Return(mockRepo)

			svc := NewSubjectService(nil, mockProvider, nil, nil, mockObserver)

			got, err := svc.GetFullProfile(ctx, tt.subjectID, tt.subjectType)

			if tt.wantErr {
				require.Error(t, err)
				// Error occurred as expected
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want.Groups, got.Groups)
				assert.Equal(t, tt.want.Roles, got.Roles)
				assert.Equal(t, tt.want.Permissions, got.Permissions)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSubjectService_List(t *testing.T) {
	tests := []struct {
		name       string
		pagination *param.PaginationParam
		filter     *param.SubjectListFilterParam
		setup      func(*repomocks.MockSubjectRepository)
		want       *model.SubjectRoles
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "Success - Returns subject list",
			pagination: &param.PaginationParam{Page: intPtr(1), Limit: intPtr(10)},
			filter:     &param.SubjectListFilterParam{SubjectID: strPtr("user-123")},
			setup: func(m *repomocks.MockSubjectRepository) {
				m.EXPECT().List(mock.Anything, mock.AnythingOfType("*param.PaginationParam"), mock.AnythingOfType("*param.SubjectListFilterParam")).Return(model.SubjectRoles{
					Items: []model.SubjectRole{
						{SubjectID: "user-123", SubjectType: "user", RoleID: 1, RoleUID: "role-1"},
					},
					Meta: model.Meta{Page: 1, Limit: 10, Total: 1},
				}, nil)
			},
			want: &model.SubjectRoles{
				Items: []model.SubjectRole{
					{SubjectID: "user-123", SubjectType: "user", RoleID: 1, RoleUID: "role-1"},
				},
				Meta: model.Meta{Page: 1, Limit: 10, Total: 1},
			},
			wantErr: false,
		},
		{
			name:       "Error - Repository failure",
			pagination: &param.PaginationParam{Page: intPtr(1), Limit: intPtr(10)},
			filter:     &param.SubjectListFilterParam{},
			setup: func(m *repomocks.MockSubjectRepository) {
				m.EXPECT().List(mock.Anything, mock.Anything, mock.Anything).Return(model.SubjectRoles{}, assert.AnError)
			},
			wantErr: true,
			errMsg:  "failed to list subjects",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			mockRepo := repomocks.NewMockSubjectRepository(t)
			mockObserver := new(mockSubjectObserver)
			mockObserver.On("OnSignal", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

			tt.setup(mockRepo)
			if !tt.wantErr {
				mockObserver.On("OnSignal", mock.Anything, signal.SignalSuccess, mock.Anything, nil).Once()
			}

			mockProvider := new(repomocks.MockRepositoryProvider)
			mockProvider.EXPECT().Subject().Return(mockRepo)

			svc := NewSubjectService(nil, mockProvider, nil, nil, mockObserver)

			got, err := svc.List(ctx, tt.pagination, tt.filter)

			if tt.wantErr {
				require.Error(t, err)
				// Error occurred as expected
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSubjectService_Assign(t *testing.T) {
	tests := []struct {
		name        string
		subjectID   string
		subjectType string
		roleUID     string
		setup       func(*repomocks.MockUnitOfWork, *repomocks.MockRepositoryProvider, *repomocks.MockRoleRepository, *repomocks.MockSubjectRepository, *eventmocks.MockEventPublisher, *mockSubjectObserver)
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "Success - Assigns role",
			subjectID:   "user-123",
			subjectType: "user",
			roleUID:     "role-1",
			setup: func(uow *repomocks.MockUnitOfWork, provider *repomocks.MockRepositoryProvider, roleRepo *repomocks.MockRoleRepository, subjectRepo *repomocks.MockSubjectRepository, pub *eventmocks.MockEventPublisher, obs *mockSubjectObserver) {
				role := &model.Role{ID: 1, UID: "role-1", Name: "Admin", GroupID: 1, GroupUID: "group-1"}

				uow.EXPECT().Do(mock.Anything, mock.AnythingOfType("func(repository.RepositoryProvider) error")).Return(nil).Run(func(ctx context.Context, fn func(repository.RepositoryProvider) error) {
					provider.EXPECT().Role().Return(roleRepo)
					roleRepo.EXPECT().GetByUID(mock.Anything, "role-1").Return(role, nil)

					provider.EXPECT().Subject().Return(subjectRepo)
					subjectRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*model.SubjectRole")).Return(nil)

					err := fn(provider)
					assert.NoError(t, err)
				})

				pub.EXPECT().Publish(mock.Anything, event.EventSubjectAssign, mock.AnythingOfType("*event.EventSubjectAssignData")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:        "Error - Role not found",
			subjectID:   "user-123",
			subjectType: "user",
			roleUID:     "nonexistent",
			setup: func(uow *repomocks.MockUnitOfWork, provider *repomocks.MockRepositoryProvider, roleRepo *repomocks.MockRoleRepository, subjectRepo *repomocks.MockSubjectRepository, pub *eventmocks.MockEventPublisher, obs *mockSubjectObserver) {
				uow.EXPECT().Do(mock.Anything, mock.AnythingOfType("func(repository.RepositoryProvider) error")).Return(assert.AnError).Run(func(ctx context.Context, fn func(repository.RepositoryProvider) error) {
					provider.EXPECT().Role().Return(roleRepo)
					roleRepo.EXPECT().GetByUID(mock.Anything, "nonexistent").Return(nil, assert.AnError)

					// Set up observer expectations first
					fn(provider)
				})
			},
			wantErr: true,
			// Error is wrapped by service
		},
		{
			name:        "Error - Create assignment fails",
			subjectID:   "user-123",
			subjectType: "user",
			roleUID:     "role-1",
			setup: func(uow *repomocks.MockUnitOfWork, provider *repomocks.MockRepositoryProvider, roleRepo *repomocks.MockRoleRepository, subjectRepo *repomocks.MockSubjectRepository, pub *eventmocks.MockEventPublisher, obs *mockSubjectObserver) {
				role := &model.Role{ID: 1, UID: "role-1", Name: "Admin", GroupID: 1, GroupUID: "group-1"}

				uow.EXPECT().Do(mock.Anything, mock.AnythingOfType("func(repository.RepositoryProvider) error")).Return(assert.AnError).Run(func(ctx context.Context, fn func(repository.RepositoryProvider) error) {
					provider.EXPECT().Role().Return(roleRepo)
					roleRepo.EXPECT().GetByUID(mock.Anything, "role-1").Return(role, nil)

					provider.EXPECT().Subject().Return(subjectRepo)
					subjectRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*model.SubjectRole")).Return(assert.AnError)

					// Set up observer expectations first
					fn(provider)
				})
			},
			wantErr: true,
			errMsg:  "failed to assign role",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			mockUow := repomocks.NewMockUnitOfWork(t)
			mockProvider := repomocks.NewMockRepositoryProvider(t)
			mockRoleRepo := repomocks.NewMockRoleRepository(t)
			mockSubjectRepo := repomocks.NewMockSubjectRepository(t)
			mockPublisher := eventmocks.NewMockEventPublisher(t)
			mockObserver := new(mockSubjectObserver)
			mockObserver.On("OnSignal", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

			tt.setup(mockUow, mockProvider, mockRoleRepo, mockSubjectRepo, mockPublisher, mockObserver)

			svc := NewSubjectService(mockUow, mockProvider, mockPublisher, nil, mockObserver)

			err := svc.Assign(ctx, tt.subjectID, tt.subjectType, tt.roleUID)

			if tt.wantErr {
				require.Error(t, err)
				// Error occurred as expected
			} else {
				require.NoError(t, err)
			}

			mockUow.AssertExpectations(t)
			mockProvider.AssertExpectations(t)
			mockRoleRepo.AssertExpectations(t)
			mockSubjectRepo.AssertExpectations(t)
			mockPublisher.AssertExpectations(t)
		})
	}
}

func TestSubjectService_Revoke(t *testing.T) {
	tests := []struct {
		name        string
		subjectID   string
		subjectType string
		roleUID     string
		setup       func(*repomocks.MockUnitOfWork, *repomocks.MockRepositoryProvider, *repomocks.MockRoleRepository, *repomocks.MockSubjectRepository, *eventmocks.MockEventPublisher, *mockSubjectObserver)
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "Success - Revokes role",
			subjectID:   "user-123",
			subjectType: "user",
			roleUID:     "role-1",
			setup: func(uow *repomocks.MockUnitOfWork, provider *repomocks.MockRepositoryProvider, roleRepo *repomocks.MockRoleRepository, subjectRepo *repomocks.MockSubjectRepository, pub *eventmocks.MockEventPublisher, obs *mockSubjectObserver) {
				role := &model.Role{ID: 1, UID: "role-1", Name: "Admin", GroupID: 1, GroupUID: "group-1"}

				uow.EXPECT().Do(mock.Anything, mock.AnythingOfType("func(repository.RepositoryProvider) error")).Return(nil).Run(func(ctx context.Context, fn func(repository.RepositoryProvider) error) {
					provider.EXPECT().Role().Return(roleRepo)
					roleRepo.EXPECT().GetByUID(mock.Anything, "role-1").Return(role, nil)

					provider.EXPECT().Subject().Return(subjectRepo)
					subjectRepo.EXPECT().Delete(mock.Anything, "user-123", "user", int64(1)).Return(nil)

					err := fn(provider)
					assert.NoError(t, err)
				})

				pub.EXPECT().Publish(mock.Anything, event.EventSubjectRevoke, mock.AnythingOfType("*event.EventSubjectRevokeData")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:        "Error - Role not found",
			subjectID:   "user-123",
			subjectType: "user",
			roleUID:     "nonexistent",
			setup: func(uow *repomocks.MockUnitOfWork, provider *repomocks.MockRepositoryProvider, roleRepo *repomocks.MockRoleRepository, subjectRepo *repomocks.MockSubjectRepository, pub *eventmocks.MockEventPublisher, obs *mockSubjectObserver) {
				uow.EXPECT().Do(mock.Anything, mock.AnythingOfType("func(repository.RepositoryProvider) error")).Return(assert.AnError).Run(func(ctx context.Context, fn func(repository.RepositoryProvider) error) {
					provider.EXPECT().Role().Return(roleRepo)
					roleRepo.EXPECT().GetByUID(mock.Anything, "nonexistent").Return(nil, assert.AnError)

					// Set up observer expectations first
					fn(provider)
				})
			},
			wantErr: true,
			// Error is wrapped by service
		},
		{
			name:        "Error - Delete fails",
			subjectID:   "user-123",
			subjectType: "user",
			roleUID:     "role-1",
			setup: func(uow *repomocks.MockUnitOfWork, provider *repomocks.MockRepositoryProvider, roleRepo *repomocks.MockRoleRepository, subjectRepo *repomocks.MockSubjectRepository, pub *eventmocks.MockEventPublisher, obs *mockSubjectObserver) {
				role := &model.Role{ID: 1, UID: "role-1", Name: "Admin", GroupID: 1, GroupUID: "group-1"}

				uow.EXPECT().Do(mock.Anything, mock.AnythingOfType("func(repository.RepositoryProvider) error")).Return(assert.AnError).Run(func(ctx context.Context, fn func(repository.RepositoryProvider) error) {
					provider.EXPECT().Role().Return(roleRepo)
					roleRepo.EXPECT().GetByUID(mock.Anything, "role-1").Return(role, nil)

					provider.EXPECT().Subject().Return(subjectRepo)
					subjectRepo.EXPECT().Delete(mock.Anything, "user-123", "user", int64(1)).Return(assert.AnError)

					// Set up observer expectations first
					fn(provider)
				})
			},
			wantErr: true,
			errMsg:  "failed to revoke role",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			mockUow := repomocks.NewMockUnitOfWork(t)
			mockProvider := repomocks.NewMockRepositoryProvider(t)
			mockRoleRepo := repomocks.NewMockRoleRepository(t)
			mockSubjectRepo := repomocks.NewMockSubjectRepository(t)
			mockPublisher := eventmocks.NewMockEventPublisher(t)
			mockObserver := new(mockSubjectObserver)
			mockObserver.On("OnSignal", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

			tt.setup(mockUow, mockProvider, mockRoleRepo, mockSubjectRepo, mockPublisher, mockObserver)

			svc := NewSubjectService(mockUow, mockProvider, mockPublisher, nil, mockObserver)

			err := svc.Revoke(ctx, tt.subjectID, tt.subjectType, tt.roleUID)

			if tt.wantErr {
				require.Error(t, err)
				// Error occurred as expected
			} else {
				require.NoError(t, err)
			}

			mockUow.AssertExpectations(t)
			mockProvider.AssertExpectations(t)
			mockRoleRepo.AssertExpectations(t)
			mockSubjectRepo.AssertExpectations(t)
			mockPublisher.AssertExpectations(t)
		})
	}
}

// Helper functions for pointer conversions
func intPtr(i int) *int {
	return &i
}

func strPtr(s string) *string {
	return &s
}
