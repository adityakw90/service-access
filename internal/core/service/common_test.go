package service

import (
	"context"

	"github.com/adityakw90/service-access/internal/core/domain/event"
	"github.com/adityakw90/service-access/internal/core/port/repository"
)

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

func (m *mockPublisher) Name() string {
	return "mock-publisher"
}

func (m *mockPublisher) Close() error {
	return nil
}
