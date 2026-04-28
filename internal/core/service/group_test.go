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

func TestGroupService_Create(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*repomocks.MockUnitOfWork, *repomocks.MockRepositoryProvider, *securitymocks.MockUIDGenerator, *eventmocks.MockEventPublisher)
		param   param.GroupCreateParam
		want    *model.Group
		wantErr bool
		errMsg  string
	}{
		{
			name: "Happy Path",
			setup: func(m *repomocks.MockUnitOfWork, p *repomocks.MockRepositoryProvider, uidGen *securitymocks.MockUIDGenerator, pub *eventmocks.MockEventPublisher) {
				uidGen.EXPECT().New().Return("test-uid")
				pub.EXPECT().Publish(mock.Anything, mock.Anything, mock.AnythingOfType("*event.EventGroupCreateData")).Return(nil)
				m.EXPECT().Do(mock.Anything, mock.AnythingOfType("func(repository.RepositoryProvider) error")).Return(nil).RunAndReturn(func(ctx context.Context, fn func(repository.RepositoryProvider) error) error {
					mockGroupRepo := repomocks.NewMockGroupRepository(t)
					mockGroupRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*model.Group")).Return(nil)
					p.EXPECT().Group().Return(mockGroupRepo)
					return fn(p)
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
			setup: func(m *repomocks.MockUnitOfWork, p *repomocks.MockRepositoryProvider, uidGen *securitymocks.MockUIDGenerator, pub *eventmocks.MockEventPublisher) {
				uidGen.EXPECT().New().Return("test-uid")
				m.EXPECT().Do(mock.Anything, mock.AnythingOfType("func(repository.RepositoryProvider) error")).Return(errors.New("transaction error"))
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
			mockPublisher := eventmocks.NewMockEventPublisher(t)
			mockUIDGenerator := securitymocks.NewMockUIDGenerator(t)
			mockResolverProvider := resolvermocks.NewMockResolverProvider(t)
			mockObserver := adapterobserver.NewNoopObserver[signal.SignalGroup]()

			tt.setup(mockUoW, mockRepos, mockUIDGenerator, mockPublisher)

			service := NewGroupService(mockUoW, mockRepos, mockPublisher, mockUIDGenerator, mockResolverProvider, nil, mockObserver, adapterobserver.NewNoopObserver[signal.SignalGroupPermission]())
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
				mockGroupRepo := repomocks.NewMockGroupRepository(t)
				mockGroupRepo.EXPECT().GetByID(mock.Anything, int64(1)).Return(&model.Group{
					ID:          1,
					UID:         "test-uid",
					Name:        "admin",
					Description: "Administrators",
				}, nil)
				p.EXPECT().Group().Return(mockGroupRepo)
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
				mockGroupRepo := repomocks.NewMockGroupRepository(t)
				mockGroupRepo.EXPECT().GetByID(mock.Anything, int64(1)).Return(nil, errors.New("group not found"))
				p.EXPECT().Group().Return(mockGroupRepo)
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
			mockPublisher := eventmocks.NewMockEventPublisher(t)
			mockUIDGenerator := securitymocks.NewMockUIDGenerator(t)
			mockResolverProvider := resolvermocks.NewMockResolverProvider(t)
			mockObserver := adapterobserver.NewNoopObserver[signal.SignalGroup]()

			// Set up resolver mock
			mockGroupResolver := resolvermocks.NewMockGroupResolver(t)
			mockGroupResolver.EXPECT().IDsByUIDs(mock.Anything, []string{tt.uid}).Return(map[string]int64{tt.uid: 1}, nil)
			mockResolverProvider.EXPECT().Group().Return(mockGroupResolver)

			tt.setup(mockRepos)

			service := NewGroupService(mockUoW, mockRepos, mockPublisher, mockUIDGenerator, mockResolverProvider, nil, mockObserver, adapterobserver.NewNoopObserver[signal.SignalGroupPermission]())
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
				mockGroupRepo := repomocks.NewMockGroupRepository(t)
				mockGroupRepo.EXPECT().List(mock.Anything, mock.Anything, mock.Anything).Return(model.Groups{
					Items: []model.Group{
						{ID: 1, UID: "test-uid-1", Name: "admin"},
						{ID: 2, UID: "test-uid-2", Name: "user"},
					},
					Meta: model.Meta{Total: 2},
				}, nil)
				p.EXPECT().Group().Return(mockGroupRepo)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUoW := repomocks.NewMockUnitOfWork(t)
			mockRepos := repomocks.NewMockRepositoryProvider(t)
			mockPublisher := eventmocks.NewMockEventPublisher(t)
			mockUIDGenerator := securitymocks.NewMockUIDGenerator(t)
			mockResolverProvider := resolvermocks.NewMockResolverProvider(t)
			mockObserver := adapterobserver.NewNoopObserver[signal.SignalGroup]()

			tt.setup(mockRepos)

			service := NewGroupService(mockUoW, mockRepos, mockPublisher, mockUIDGenerator, mockResolverProvider, nil, mockObserver, adapterobserver.NewNoopObserver[signal.SignalGroupPermission]())
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
