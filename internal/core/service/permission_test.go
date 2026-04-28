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

func TestPermissionService_Create(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*repomocks.MockUnitOfWork, *repomocks.MockRepositoryProvider, *securitymocks.MockUIDGenerator)
		param   param.PermissionCreateParam
		want    *model.Permission
		wantErr bool
		errMsg  string
	}{
		{
			name: "Happy Path",
			setup: func(m *repomocks.MockUnitOfWork, p *repomocks.MockRepositoryProvider, uidGen *securitymocks.MockUIDGenerator) {
				uidGen.EXPECT().New().Return("test-uid")
				m.EXPECT().Do(mock.Anything, mock.AnythingOfType("func(repository.RepositoryProvider) error")).Return(nil).RunAndReturn(func(ctx context.Context, fn func(repository.RepositoryProvider) error) error {
					mockPermRepo := repomocks.NewMockPermissionRepository(t)
					repos := &mockRepositories{}
					repos.SetPermissionRepo(mockPermRepo)

					// Set up the mock to return nil for Create
					mockPermRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*model.Permission")).Return(nil)

					// Call the function with our mock repositories
					return fn(repos)
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
			setup: func(m *repomocks.MockUnitOfWork, p *repomocks.MockRepositoryProvider, uidGen *securitymocks.MockUIDGenerator) {
				// Don't set up UID generator expectation since it won't be called in error case
				m.EXPECT().Do(mock.Anything, mock.AnythingOfType("func(repository.RepositoryProvider) error")).Return(errors.New("transaction error"))
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
			mockUoW := repomocks.NewMockUnitOfWork(t)
			mockRepos := repomocks.NewMockRepositoryProvider(t)
			mockPublisher := &mockPublisher{}
			mockUIDGenerator := securitymocks.NewMockUIDGenerator(t)
			mockResolverProvider := resolvermocks.NewMockResolverProvider(t)
			mockObserver := adapterobserver.NewNoopObserver[signal.SignalPermission]()

			tt.setup(mockUoW, mockRepos, mockUIDGenerator)

			service := NewPermissionService(mockUoW, mockRepos, mockPublisher, mockUIDGenerator, mockResolverProvider, nil, mockObserver)
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
		setup   func(*repomocks.MockRepositoryProvider)
		uid     string
		want    *model.Permission
		wantErr bool
		errMsg  string
	}{
		{
			name: "Happy Path",
			setup: func(p *repomocks.MockRepositoryProvider) {
				mockPermRepo := repomocks.NewMockPermissionRepository(t)
				mockPermRepo.EXPECT().GetByID(mock.Anything, int64(1)).Return(&model.Permission{
					ID:          1,
					UID:         "test-uid",
					Resource:    "articles",
					Action:      "read",
					Description: "Read articles",
				}, nil)
				p.EXPECT().Permission().Return(mockPermRepo)
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
			setup: func(p *repomocks.MockRepositoryProvider) {
				mockPermRepo := repomocks.NewMockPermissionRepository(t)
				mockPermRepo.EXPECT().GetByID(mock.Anything, int64(1)).Return(nil, errors.New("permission not found"))
				p.EXPECT().Permission().Return(mockPermRepo)
			},
			uid:     "not-found",
			wantErr: true,
			errMsg:  "failed to get permission",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUoW := repomocks.NewMockUnitOfWork(t)
			mockRepos := repomocks.NewMockRepositoryProvider(t)
			mockPublisher := &mockPublisher{}
			mockUIDGenerator := securitymocks.NewMockUIDGenerator(t)
			mockResolverProvider := resolvermocks.NewMockResolverProvider(t)
			mockObserver := adapterobserver.NewNoopObserver[signal.SignalPermission]()

			// Set up resolver mock
			mockPermissionResolver := resolvermocks.NewMockPermissionResolver(t)
			mockPermissionResolver.EXPECT().IDsByUIDs(mock.Anything, []string{tt.uid}).Return(map[string]int64{tt.uid: 1}, nil)
			mockResolverProvider.EXPECT().Permission().Return(mockPermissionResolver)

			tt.setup(mockRepos)

			service := NewPermissionService(mockUoW, mockRepos, mockPublisher, mockUIDGenerator, mockResolverProvider, nil, mockObserver)
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
		setup   func(*repomocks.MockRepositoryProvider)
		wantErr bool
	}{
		{
			name: "Happy Path",
			setup: func(p *repomocks.MockRepositoryProvider) {
				mockPermRepo := repomocks.NewMockPermissionRepository(t)
				mockPermRepo.EXPECT().List(mock.Anything, mock.Anything, mock.Anything).Return(model.Permissions{
					Items: []model.Permission{
						{ID: 1, UID: "test-uid-1", Resource: "articles", Action: "read"},
						{ID: 2, UID: "test-uid-2", Resource: "articles", Action: "write"},
					},
					Meta: model.Meta{Total: 2},
				}, nil)
				p.EXPECT().Permission().Return(mockPermRepo)
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
			mockObserver := adapterobserver.NewNoopObserver[signal.SignalPermission]()

			tt.setup(mockRepos)

			service := NewPermissionService(mockUoW, mockRepos, mockPublisher, mockUIDGenerator, mockResolverProvider, nil, mockObserver)
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
