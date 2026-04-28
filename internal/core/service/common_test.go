package service

import (
	"context"

	"github.com/adityakw90/service-access/internal/core/domain/event"
	"github.com/adityakw90/service-access/internal/core/port/repository"
	repomocks "github.com/adityakw90/service-access/mocks/repository"
)

// mockRepositories implements repository.RepositoryProvider for testing using generated mocks
type mockRepositories struct {
	group      repository.GroupRepository
	permission repository.PermissionRepository
	role       repository.RoleRepository
	subject    repository.SubjectRepository
}

func newMockRepositories() *mockRepositories {
	return &mockRepositories{}
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

// Helper functions to set up individual repository mocks
func (m *mockRepositories) SetGroupRepo(repo *repomocks.MockGroupRepository) {
	m.group = repo
}

func (m *mockRepositories) SetPermissionRepo(repo *repomocks.MockPermissionRepository) {
	m.permission = repo
}

func (m *mockRepositories) SetRoleRepo(repo *repomocks.MockRoleRepository) {
	m.role = repo
}

func (m *mockRepositories) SetSubjectRepo(repo *repomocks.MockSubjectRepository) {
	m.subject = repo
}

// mockPublisher is a simple mock for event.Publisher for testing event publishing
type mockPublisher struct {
	publishedEvents []publishedEvent
}

type publishedEvent struct {
	eventType event.EventType
	eventData any
}

func newMockPublisher() *mockPublisher {
	return &mockPublisher{publishedEvents: make([]publishedEvent, 0)}
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

func (m *mockPublisher) GetPublishedEvents() []publishedEvent {
	return m.publishedEvents
}

func (m *mockPublisher) ClearEvents() {
	m.publishedEvents = make([]publishedEvent, 0)
}
