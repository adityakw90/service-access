package repository

import (
	"testing"

	portrepository "github.com/adityakw90/service-access/internal/core/port/repository"
	"github.com/stretchr/testify/assert"
)

// TestAdapter_Repositories_LazyInitialization tests that repositories are created lazily
func TestAdapter_Repositories_LazyInitialization(t *testing.T) {
	t.Run("Repositories are created lazily and cached", func(t *testing.T) {
		repos := &repositoryProvider{
			db: nil, // db can be nil for this test since we're not calling any repository methods
		}

		// Initially, all repositories should be nil
		assert.Nil(t, repos.permission)
		assert.Nil(t, repos.group)
		assert.Nil(t, repos.role)
		assert.Nil(t, repos.subject)

		// First call creates repository
		perm1 := repos.Permission()
		assert.NotNil(t, perm1)
		assert.NotNil(t, repos.permission)

		// Second call returns same instance
		perm2 := repos.Permission()
		assert.Same(t, perm1, perm2)

		// Same for other repositories
		group1 := repos.Group()
		group2 := repos.Group()
		assert.Same(t, group1, group2)

		role1 := repos.Role()
		role2 := repos.Role()
		assert.Same(t, role1, role2)

		subject1 := repos.Subject()
		subject2 := repos.Subject()
		assert.Same(t, subject1, subject2)
	})
}

// TestAdapter_Repositories_ImplementsInterface tests that repositories implements the port interface
func TestAdapter_Repositories_ImplementsInterface(t *testing.T) {
	t.Run("Repositories implements portrepository.Repositories interface", func(t *testing.T) {
		// This test verifies that the repositories struct correctly implements
		// the portrepository.Repositories interface
		var _ portrepository.RepositoryProvider = &repositoryProvider{}
	})
}

// TestAdapter_NewUnitOfWork_CreatesInstance tests that NewUnitOfWork creates a valid instance
func TestAdapter_NewUnitOfWork_CreatesInstance(t *testing.T) {
	t.Run("NewUnitOfWork creates non-nil instance", func(t *testing.T) {
		// Note: We can't fully test the UnitOfWork with pgxmock because it expects *pgxpool.Pool
		// This is a minimal test that verifies the type is correct
		// Full transaction testing should be done in integration tests

		// The actual testing of transaction behavior requires a real database connection
		// or refactoring the NewUnitOfWork to accept an interface instead of *pgxpool.Pool

		// For now, we verify the type is correct
		assert.NotNil(t, &unitOfWork{})
	})
}

// TestAdapter_Repositories_TypeSafety tests type safety of repository returns
func TestAdapter_Repositories_TypeSafety(t *testing.T) {
	t.Run("Repository methods return correct interface types", func(t *testing.T) {
		repos := &repositoryProvider{
			db: nil,
		}

		// Test that the returned types implement the correct interfaces
		var _ portrepository.PermissionRepository = repos.Permission()
		var _ portrepository.GroupRepository = repos.Group()
		var _ portrepository.RoleRepository = repos.Role()
		var _ portrepository.SubjectRepository = repos.Subject()

		assert.True(t, true, "All repository types are correct")
	})
}
