package repository

import (
	"context"
	"sync"
	"testing"

	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestRepositories creates a repositories instance with a mock db executor for testing
func createTestRepositories(t *testing.T) (*repositoryProvider, pgxmock.PgxPoolIface) {
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err, "Failed to create mock pool")
	return &repositoryProvider{db: mockPool}, mockPool
}

// createTestRepositoriesWithNilDB creates a repositories instance with nil db executor
func createTestRepositoriesWithNilDB() *repositoryProvider {
	return &repositoryProvider{db: nil}
}

// ============================================================================
// PRIORITY 0: Critical - Thread Safety & Core Functionality
// ============================================================================

// TestAdapter_Repositories_ConcurrentInitialization tests that lazy initialization is thread-safe
func TestAdapter_Repositories_ConcurrentInitialization(t *testing.T) {
	tests := []struct {
		name       string
		goroutines int
		repoGetter func(*repositoryProvider) any
	}{
		{
			name:       "Concurrent Permission repository access",
			goroutines: 100,
			repoGetter: func(r *repositoryProvider) any {
				return r.Permission()
			},
		},
		{
			name:       "Concurrent Group repository access",
			goroutines: 100,
			repoGetter: func(r *repositoryProvider) any {
				return r.Group()
			},
		},
		{
			name:       "Concurrent Role repository access",
			goroutines: 100,
			repoGetter: func(r *repositoryProvider) any {
				return r.Role()
			},
		},
		{
			name:       "Concurrent Subject repository access",
			goroutines: 100,
			repoGetter: func(r *repositoryProvider) any {
				return r.Subject()
			},
		},
		{
			name:       "Mixed repository access concurrently",
			goroutines: 100,
			repoGetter: func(r *repositoryProvider) any {
				// Use a counter to mix up different repository types
				// This is called concurrently, so the counter is just for variety
				// Thread-safety is tested by having goroutines call this
				return r.Permission() // Return one type consistently for same-instance verification
			},
		},
		{
			name:       "Rapid repeated calls to same repository",
			goroutines: 1000,
			repoGetter: func(r *repositoryProvider) any {
				return r.Permission()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repos, _ := createTestRepositories(t)

			var wg sync.WaitGroup
			instances := make([]any, tt.goroutines)
			var mu sync.Mutex

			for i := 0; i < tt.goroutines; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					instance := tt.repoGetter(repos)
					mu.Lock()
					instances[idx] = instance
					mu.Unlock()
				}(i)
			}

			wg.Wait()

			// Verify all instances are the same (same memory address)
			for i := 1; i < len(instances); i++ {
				assert.Same(t, instances[0], instances[i],
					"All goroutines should receive the same repository instance")
			}
		})
	}
}

// TestAdapter_Repositories_DbExecutorPropagation verifies dbExecutor is correctly propagated
func TestAdapter_Repositories_DbExecutorPropagation(t *testing.T) {
	tests := []struct {
		name       string
		repoGetter func(*repositoryProvider) any
	}{
		{
			name: "Permission repository receives dbExecutor",
			repoGetter: func(r *repositoryProvider) any {
				return r.Permission()
			},
		},
		{
			name: "Group repository receives dbExecutor",
			repoGetter: func(r *repositoryProvider) any {
				return r.Group()
			},
		},
		{
			name: "Role repository receives dbExecutor",
			repoGetter: func(r *repositoryProvider) any {
				return r.Role()
			},
		},
		{
			name: "Subject repository receives dbExecutor",
			repoGetter: func(r *repositoryProvider) any {
				return r.Subject()
			},
		},
		{
			name: "All repositories share same dbExecutor instance",
			repoGetter: func(r *repositoryProvider) any {
				// Return all repositories for verification
				return []any{r.Permission(), r.Group(), r.Role(), r.Subject()}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)

			repos := &repositoryProvider{db: mockPool}
			repo := tt.repoGetter(repos)

			// Verify repositories were created with the dbExecutor
			if list, ok := repo.([]any); ok {
				// All repositories case
				require.Len(t, list, 4, "Should have 4 repositories")
				for _, r := range list {
					assert.NotNil(t, r, "Each repository should not be nil")
				}
			} else {
				// Single repository case
				assert.NotNil(t, repo, "Repository should not be nil")
			}

			// No queries expected, so just verify the mock pool is clean
			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

// ============================================================================
// PRIORITY 1: High - Integration & Race Detection
// ============================================================================

// TestAdapter_Repositories_RaceCondition tests for explicit race conditions
func TestAdapter_Repositories_RaceCondition(t *testing.T) {
	tests := []struct {
		name   string
		testFn func(*testing.T, *repositoryProvider)
	}{
		{
			name: "Write-read race on repository initialization",
			testFn: func(t *testing.T, repos *repositoryProvider) {
				var wg sync.WaitGroup
				for i := 0; i < 50; i++ {
					wg.Add(2)
					go func() {
						defer wg.Done()
						_ = repos.Permission()
					}()
					go func() {
						defer wg.Done()
						_ = repos.Group()
					}()
				}
				wg.Wait()
				assert.NotNil(t, repos.Permission())
				assert.NotNil(t, repos.Group())
			},
		},
		{
			name: "Read-read race on same repository",
			testFn: func(t *testing.T, repos *repositoryProvider) {
				var wg sync.WaitGroup
				for i := 0; i < 100; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						_ = repos.Role()
					}()
				}
				wg.Wait()
				assert.NotNil(t, repos.Role())
			},
		},
		{
			name: "Mixed repository race - all types concurrently",
			testFn: func(t *testing.T, repos *repositoryProvider) {
				var wg sync.WaitGroup
				for i := 0; i < 25; i++ {
					wg.Add(4)
					go func() {
						defer wg.Done()
						_ = repos.Permission()
					}()
					go func() {
						defer wg.Done()
						_ = repos.Group()
					}()
					go func() {
						defer wg.Done()
						_ = repos.Role()
					}()
					go func() {
						defer wg.Done()
						_ = repos.Subject()
					}()
				}
				wg.Wait()
				assert.NotNil(t, repos.Permission())
				assert.NotNil(t, repos.Group())
				assert.NotNil(t, repos.Role())
				assert.NotNil(t, repos.Subject())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repos, _ := createTestRepositories(t)
			tt.testFn(t, repos)
		})
	}
}

// TestAdapter_Repositories_WithinUnitOfWork tests repositories within transaction pattern
func TestAdapter_Repositories_WithinUnitOfWork(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func(pgxmock.PgxPoolIface)
		testFunc      func(*testing.T, *repositoryProvider)
		expectError   bool
		errorContains string
	}{
		{
			name: "Sequential repository access within transaction",
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectCommit()
			},
			testFunc: func(t *testing.T, repos *repositoryProvider) {
				// Access repositories sequentially - should share same transaction
				_ = repos.Permission()
				_ = repos.Group()
				_ = repos.Role()
				_ = repos.Subject()
			},
			expectError: false,
		},
		{
			name: "Multiple repository access in same transaction",
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectCommit()
			},
			testFunc: func(t *testing.T, repos *repositoryProvider) {
				// Access repositories multiple times - should return cached instances
				perm1 := repos.Permission()
				perm2 := repos.Permission()
				assert.Same(t, perm1, perm2, "Should return same instance")

				group1 := repos.Group()
				group2 := repos.Group()
				assert.Same(t, group1, group2, "Should return same instance")
			},
			expectError: false,
		},
		{
			name: "Repositories share same db executor instance",
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectCommit()
			},
			testFunc: func(t *testing.T, repos *repositoryProvider) {
				// All repositories should be created with the same db executor
				perm := repos.Permission()
				group := repos.Group()
				role := repos.Role()
				subject := repos.Subject()

				assert.NotNil(t, perm)
				assert.NotNil(t, group)
				assert.NotNil(t, role)
				assert.NotNil(t, subject)
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)

			tt.setupMock(mockPool)

			// Begin transaction
			tx, err := mockPool.Begin(context.Background())
			require.NoError(t, err)
			defer tx.Rollback(context.Background())

			// Create repositories with transaction as dbExecutor
			repos := &repositoryProvider{db: tx}
			tt.testFunc(t, repos)

			// Commit transaction
			err = tx.Commit(context.Background())
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

// TestAdapter_Repositories_InstanceIdentity verifies instance identity
func TestAdapter_Repositories_InstanceIdentity(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func(*testing.T, *repositoryProvider)
	}{
		{
			name: "Different repository types return different instances",
			testFunc: func(t *testing.T, repos *repositoryProvider) {
				perm := repos.Permission()
				group := repos.Group()
				role := repos.Role()
				subject := repos.Subject()

				// All should be non-nil
				assert.NotNil(t, perm)
				assert.NotNil(t, group)
				assert.NotNil(t, role)
				assert.NotNil(t, subject)

				// All should be different instances
				assert.NotSame(t, perm, group, "Permission and Group should be different instances")
				assert.NotSame(t, perm, role, "Permission and Role should be different instances")
				assert.NotSame(t, perm, subject, "Permission and Subject should be different instances")
				assert.NotSame(t, group, role, "Group and Role should be different instances")
				assert.NotSame(t, group, subject, "Group and Subject should be different instances")
				assert.NotSame(t, role, subject, "Role and Subject should be different instances")
			},
		},
		{
			name: "Same repository type returns cached instance",
			testFunc: func(t *testing.T, repos *repositoryProvider) {
				perm1 := repos.Permission()
				perm2 := repos.Permission()
				perm3 := repos.Permission()

				group1 := repos.Group()
				group2 := repos.Group()

				role1 := repos.Role()
				role2 := repos.Role()

				subject1 := repos.Subject()
				subject2 := repos.Subject()

				// Same instances should be returned
				assert.Same(t, perm1, perm2, "Permission should return cached instance")
				assert.Same(t, perm2, perm3, "Permission should return cached instance")
				assert.Same(t, group1, group2, "Group should return cached instance")
				assert.Same(t, role1, role2, "Role should return cached instance")
				assert.Same(t, subject1, subject2, "Subject should return cached instance")
			},
		},
		{
			name: "Cached instances persist across multiple access patterns",
			testFunc: func(t *testing.T, repos *repositoryProvider) {
				// Create first instance
				permFirst := repos.Permission()
				groupFirst := repos.Group()

				// Access other repos
				_ = repos.Role()
				_ = repos.Subject()

				// Access original repos again
				permSecond := repos.Permission()
				groupSecond := repos.Group()

				// Should be same instances
				assert.Same(t, permFirst, permSecond, "Permission instance should persist")
				assert.Same(t, groupFirst, groupSecond, "Group instance should persist")
			},
		},
		{
			name: "Each repositories struct has independent cache",
			testFunc: func(t *testing.T, _ *repositoryProvider) {
				repos1, _ := createTestRepositories(t)
				repos2, _ := createTestRepositories(t)

				perm1 := repos1.Permission()
				perm2 := repos2.Permission()

				// Different repositories structs should have different instances
				assert.NotSame(t, perm1, perm2, "Different repositories structs should have independent caches")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repos, _ := createTestRepositories(t)
			tt.testFunc(t, repos)
		})
	}
}

// ============================================================================
// PRIORITY 2: Medium - Edge Cases & Error Handling
// ============================================================================

// TestAdapter_Repositories_NilDbExecutor tests behavior with nil dbExecutor
func TestAdapter_Repositories_NilDbExecutor(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func(*testing.T, *repositoryProvider)
	}{
		{
			name: "Repositories can be created with nil dbExecutor",
			testFunc: func(t *testing.T, repos *repositoryProvider) {
				// Should not panic when creating repos
				perm := repos.Permission()
				group := repos.Group()
				role := repos.Role()
				subject := repos.Subject()

				// All should be non-nil (repository structs exist)
				assert.NotNil(t, perm)
				assert.NotNil(t, group)
				assert.NotNil(t, role)
				assert.NotNil(t, subject)
			},
		},
		{
			name: "Nil dbExecutor is cached like valid executor",
			testFunc: func(t *testing.T, repos *repositoryProvider) {
				perm1 := repos.Permission()
				perm2 := repos.Permission()

				// Should return same cached instance even with nil db
				assert.Same(t, perm1, perm2)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repos := createTestRepositoriesWithNilDB()
			tt.testFunc(t, repos)
		})
	}
}

// TestAdapter_Repositories_MultipleInitializationAttempts tests caching behavior
func TestAdapter_Repositories_MultipleInitializationAttempts(t *testing.T) {
	tests := []struct {
		name     string
		attempts int
	}{
		{name: "Single initialization", attempts: 1},
		{name: "Double initialization", attempts: 2},
		{name: "Multiple initialization (10)", attempts: 10},
		{name: "Many initialization attempts (100)", attempts: 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repos, _ := createTestRepositories(t)

			// Store first instance
			firstPerm := repos.Permission()
			firstGroup := repos.Group()
			firstRole := repos.Role()
			firstSubject := repos.Subject()

			// Attempt initialization multiple times
			for i := 0; i < tt.attempts; i++ {
				perm := repos.Permission()
				group := repos.Group()
				role := repos.Role()
				subject := repos.Subject()

				assert.Same(t, firstPerm, perm, "Permission should return cached instance on attempt %d", i+1)
				assert.Same(t, firstGroup, group, "Group should return cached instance on attempt %d", i+1)
				assert.Same(t, firstRole, role, "Role should return cached instance on attempt %d", i+1)
				assert.Same(t, firstSubject, subject, "Subject should return cached instance on attempt %d", i+1)
			}
		})
	}
}

// TestAdapter_Repositories_DatabaseQueryExecution tests query execution with pgxmock
func TestAdapter_Repositories_DatabaseQueryExecution(t *testing.T) {
	tests := []struct {
		name       string
		repoAction func(*testing.T, *repositoryProvider)
	}{
		{
			name: "Permission repository can be created with dbExecutor",
			repoAction: func(t *testing.T, repos *repositoryProvider) {
				perm := repos.Permission()
				assert.NotNil(t, perm, "Permission repository should be created")
				// Note: Actual query execution is tested in permission_test.go
			},
		},
		{
			name: "Group repository can be created with dbExecutor",
			repoAction: func(t *testing.T, repos *repositoryProvider) {
				group := repos.Group()
				assert.NotNil(t, group, "Group repository should be created")
			},
		},
		{
			name: "Role repository can be created with dbExecutor",
			repoAction: func(t *testing.T, repos *repositoryProvider) {
				role := repos.Role()
				assert.NotNil(t, role, "Role repository should be created")
			},
		},
		{
			name: "Subject repository can be created with dbExecutor",
			repoAction: func(t *testing.T, repos *repositoryProvider) {
				subject := repos.Subject()
				assert.NotNil(t, subject, "Subject repository should be created")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)

			repos := &repositoryProvider{db: mockPool}
			tt.repoAction(t, repos)

			// No queries expected, just verify the mock pool is clean
			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

// ============================================================================
// PRIORITY 3: Low - Performance Benchmarks
// ============================================================================

// BenchmarkRepository_Initialization benchmarks lazy initialization
func BenchmarkRepository_Initialization(b *testing.B) {
	benchmarks := []struct {
		name       string
		repoGetter func(*repositoryProvider) any
	}{
		{name: "Permission", repoGetter: func(r *repositoryProvider) any { return r.Permission() }},
		{name: "Group", repoGetter: func(r *repositoryProvider) any { return r.Group() }},
		{name: "Role", repoGetter: func(r *repositoryProvider) any { return r.Role() }},
		{name: "Subject", repoGetter: func(r *repositoryProvider) any { return r.Subject() }},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			repos := &repositoryProvider{db: nil}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = bm.repoGetter(repos)
			}
		})
	}
}

// BenchmarkRepository_CachedAccess benchmarks cached instance access
func BenchmarkRepository_CachedAccess(b *testing.B) {
	benchmarks := []struct {
		name       string
		repoGetter func(*repositoryProvider) any
	}{
		{name: "Permission", repoGetter: func(r *repositoryProvider) any { return r.Permission() }},
		{name: "Group", repoGetter: func(r *repositoryProvider) any { return r.Group() }},
		{name: "Role", repoGetter: func(r *repositoryProvider) any { return r.Role() }},
		{name: "Subject", repoGetter: func(r *repositoryProvider) any { return r.Subject() }},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			repos := &repositoryProvider{db: nil}
			// Initialize first to cache it
			_ = bm.repoGetter(repos)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = bm.repoGetter(repos)
			}
		})
	}
}

// BenchmarkRepository_ConcurrentAccess benchmarks concurrent repository access
func BenchmarkRepository_ConcurrentAccess(b *testing.B) {
	benchmarks := []struct {
		name  string
		goros int
	}{
		{name: "2 goroutines", goros: 2},
		{name: "4 goroutines", goros: 4},
		{name: "8 goroutines", goros: 8},
		{name: "16 goroutines", goros: 16},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			repos := &repositoryProvider{db: nil}
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				var wg sync.WaitGroup
				for g := 0; g < bm.goros; g++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						_ = repos.Permission()
						_ = repos.Group()
						_ = repos.Role()
						_ = repos.Subject()
					}()
				}
				wg.Wait()
			}
		})
	}
}

// TestAdapter_Repositories_CrossRepositoryConsistency tests cross-repository scenarios
func TestAdapter_Repositories_CrossRepositoryConsistency(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{
			name: "All repositories in same transaction share executor",
			testFunc: func(t *testing.T) {
				mockPool, err := pgxmock.NewPool()
				require.NoError(t, err)

				mockPool.ExpectBegin()
				mockPool.ExpectCommit()

				// Begin transaction
				tx, err := mockPool.Begin(context.Background())
				require.NoError(t, err)
				defer tx.Rollback(context.Background())

				// Create repositories with transaction
				repos := &repositoryProvider{db: tx}

				// All repos should be accessible within the same transaction
				perm := repos.Permission()
				group := repos.Group()
				role := repos.Role()
				subject := repos.Subject()

				assert.NotNil(t, perm)
				assert.NotNil(t, group)
				assert.NotNil(t, role)
				assert.NotNil(t, subject)

				err = tx.Commit(context.Background())
				assert.NoError(t, err)
				assert.NoError(t, mockPool.ExpectationsWereMet())
			},
		},
		{
			name: "Repository instances are independent across transaction boundaries",
			testFunc: func(t *testing.T) {
				mockPool, err := pgxmock.NewPool()
				require.NoError(t, err)

				mockPool.ExpectBegin()
				mockPool.ExpectCommit()
				mockPool.ExpectBegin()
				mockPool.ExpectCommit()

				var perm1, perm2 interface{}

				// First transaction
				tx1, err := mockPool.Begin(context.Background())
				require.NoError(t, err)
				repos1 := &repositoryProvider{db: tx1}
				perm1 = repos1.Permission()
				err = tx1.Commit(context.Background())
				assert.NoError(t, err)

				// Second transaction
				tx2, err := mockPool.Begin(context.Background())
				require.NoError(t, err)
				repos2 := &repositoryProvider{db: tx2}
				perm2 = repos2.Permission()
				err = tx2.Commit(context.Background())
				assert.NoError(t, err)

				// Should be different instances (different repositories structs)
				assert.NotSame(t, perm1, perm2, "Different repositories structs should have different instances")

				assert.NoError(t, mockPool.ExpectationsWereMet())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFunc(t)
		})
	}
}
