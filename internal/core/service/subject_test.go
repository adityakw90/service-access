package service

import (
    "context"
    "testing"

    "github.com/adityakw90/service-access/internal/core/domain/model"
    "github.com/adityakw90/service-access/internal/core/domain/param"
    "github.com/adityakw90/service-access/internal/core/domain/signal"
    "github.com/adityakw90/service-access/internal/core/port/repository"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/require"
)

type mockSubjectRepository struct {
    mock.Mock
}

func (m *mockSubjectRepository) GetAllRoles(ctx context.Context, subjectID string, subjectType string) ([]model.Role, error) {
    args := m.Called(ctx, subjectID, subjectType)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).([]model.Role), args.Error(1)
}
func (m *mockSubjectRepository) GetAllGroups(ctx context.Context, subjectID string, subjectType string) ([]model.Group, error) {
    args := m.Called(ctx, subjectID, subjectType)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).([]model.Group), args.Error(1)
}
func (m *mockSubjectRepository) GetAllPermissions(ctx context.Context, subjectID string, subjectType string) ([]model.Permission, error) {
    args := m.Called(ctx, subjectID, subjectType)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).([]model.Permission), args.Error(1)
}

// Minimal stubs for other interface methods (not used in new tests)
func (m *mockSubjectRepository) Create(ctx context.Context, subject *model.SubjectRole) error { return nil }
func (m *mockSubjectRepository) Update(ctx context.Context, subject *model.SubjectRole) error { return nil }
func (m *mockSubjectRepository) Delete(ctx context.Context, subjectID string, subjectType string, roleID int64) error { return nil }
func (m *mockSubjectRepository) List(ctx context.Context, pagination *param.PaginationParam, filter *param.SubjectListFilterParam) (model.SubjectRoles, error) { return model.SubjectRoles{}, nil }
func (m *mockSubjectRepository) GetRoles(ctx context.Context, subjectID string, subjectType string) ([]model.SubjectRole, error) { return nil, nil }

type mockObserver struct {
    mock.Mock
}
func (m *mockObserver) OnSignal(ctx context.Context, signal signal.SignalType, data signal.SignalSubject, err error) {
    m.Called(ctx, signal, data, err)
}

type mockUnitOfWork struct{}
func (m *mockUnitOfWork) Do(ctx context.Context, fn func(repository.RepositoryProvider) error) error {
    // Provide a minimal implementation that returns a mock provider
    return fn(&mockRepositoryProvider{})
}

type mockRepositoryProvider struct{}
func (m *mockRepositoryProvider) Subject() repository.SubjectRepository {
    // Return a fresh mock with no expectations
    return &mockSubjectRepository{}
}
func (m *mockRepositoryProvider) Role() repository.RoleRepository    { return nil }
func (m *mockRepositoryProvider) Group() repository.GroupRepository  { return nil }
func (m *mockRepositoryProvider) Permission() repository.PermissionRepository { return nil }

func TestSubjectService_GetRoles(t *testing.T) {
    ctx := context.Background()
    uow := &mockUnitOfWork{}
    repos := &mockRepositoryProvider{}
    observer := &mockObserver{}
    svc := NewSubjectService(uow, repos, nil, observer)

    _, err := svc.GetRoles(ctx, "", "user")
    require.Error(t, err)
    assert.Contains(t, err.Error(), "subject_id is required")
}

func TestSubjectService_GetGroups(t *testing.T) {
    ctx := context.Background()
    uow := &mockUnitOfWork{}
    repos := &mockRepositoryProvider{}
    observer := &mockObserver{}
    svc := NewSubjectService(uow, repos, nil, observer)

    _, err := svc.GetGroups(ctx, "user-123", "")
    require.Error(t, err)
    assert.Contains(t, err.Error(), "subject_type is required")
}

func TestSubjectService_GetPermissions(t *testing.T) {
    ctx := context.Background()
    uow := &mockUnitOfWork{}
    repos := &mockRepositoryProvider{}
    observer := &mockObserver{}
    svc := NewSubjectService(uow, repos, nil, observer)

    _, err := svc.GetPermissions(ctx, "", "user")
    require.Error(t, err)
    assert.Contains(t, err.Error(), "subject_id is required")
}

func TestSubjectService_GetFullProfile_Validation(t *testing.T) {
    ctx := context.Background()
    uow := &mockUnitOfWork{}
    repos := &mockRepositoryProvider{}
    observer := &mockObserver{}
    svc := NewSubjectService(uow, repos, nil, observer)

    _, err := svc.GetFullProfile(ctx, "", "user")
    require.Error(t, err)
}
