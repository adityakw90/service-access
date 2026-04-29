package service

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/adityakw90/service-access/internal/adapter/repository"
	"github.com/adityakw90/service-access/internal/config"
	"github.com/adityakw90/service-access/internal/core/domain/model"
	"github.com/adityakw90/service-access/internal/core/domain/signal"
	portrepository "github.com/adityakw90/service-access/internal/core/port/repository"
	"github.com/adityakw90/service-access/internal/infra"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

// BenchmarkGetAllRoles benchmarks the GetAllRoles repository method
func BenchmarkGetAllRoles(b *testing.B) {
	db := getDB(b)
	ctx := context.Background()
	subjectRepo := repository.NewSubjectRepository(db)

	subjectID, subjectType := createTestData(b, ctx, db, 100, 8, 200)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := subjectRepo.GetAllRoles(ctx, subjectID, subjectType)
		if err != nil {
			b.Fatalf("GetAllRoles failed: %v", err)
		}
	}

	truncateDB(b, db)
}

// BenchmarkGetAllGroups benchmarks the GetAllGroups repository method
func BenchmarkGetAllGroups(b *testing.B) {
	db := getDB(b)
	ctx := context.Background()
	subjectRepo := repository.NewSubjectRepository(db)

	subjectID, subjectType := createTestData(b, ctx, db, 100, 10, 200)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := subjectRepo.GetAllGroups(ctx, subjectID, subjectType)
		if err != nil {
			b.Fatalf("GetAllGroups failed: %v", err)
		}
	}

	truncateDB(b, db)
}

// BenchmarkGetAllPermissions benchmarks the GetAllPermissions repository method
func BenchmarkGetAllPermissions(b *testing.B) {
	db := getDB(b)
	ctx := context.Background()
	subjectRepo := repository.NewSubjectRepository(db)

	subjectID, subjectType := createTestData(b, ctx, db, 100, 8, 200)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := subjectRepo.GetAllPermissions(ctx, subjectID, subjectType)
		if err != nil {
			b.Fatalf("GetAllPermissions failed: %v", err)
		}
	}

	truncateDB(b, db)
}

// BenchmarkGetFullProfile benchmarks the service GetFullProfile with aggregated data
func BenchmarkGetFullProfile(b *testing.B) {
	db := getDB(b)
	ctx := context.Background()

	subjectID, subjectType := createTestData(b, ctx, db, 100, 8, 200)

	// Build service with custom provider
	uow := &mockUnitOfWork{}
	repos := newCustomRepositoryProvider(db)
	observer := &mockObserver{}
	svc := NewSubjectService(uow, repos, nil, nil, observer)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := svc.GetFullProfile(ctx, subjectID, subjectType)
		if err != nil {
			b.Fatalf("GetFullProfile failed: %v", err)
		}
	}

	truncateDB(b, db)
}

// --- Data creation ---

func createTestData(
	b *testing.B,
	ctx context.Context,
	db *pgxpool.Pool,
	numRoles, numGroups, numPermissions int,
) (subjectID, subjectType string) {
	b.Helper()

	now := time.Now().UnixNano()
	subjectID = fmt.Sprintf("benchmark-subject-%d", now)
	subjectType = "user"

	// 1. Create permissions
	permRepo := repository.NewPermissionRepository(db)
	permissions := make([]*model.Permission, numPermissions)
	for i := 0; i < numPermissions; i++ {
		perm := &model.Permission{
			UID:         uuid.New().String(),
			Resource:    fmt.Sprintf("resource-%d", i),
			Action:      fmt.Sprintf("action-%d", i%5),
			Description: fmt.Sprintf("Permission %d", i),
		}
		err := permRepo.Create(ctx, perm)
		require.NoError(b, err, "failed to create permission %d", i)
		permissions[i] = perm
	}

	// 2. Create groups
	groupRepo := repository.NewGroupRepository(db)
	groups := make([]*model.Group, numGroups)
	for g := 0; g < numGroups; g++ {
		group := &model.Group{
			UID:         uuid.New().String(),
			Name:        fmt.Sprintf("benchmark-group-%d-%d", now, g),
			Description: fmt.Sprintf("Benchmark group %d", g),
		}
		err := groupRepo.Create(ctx, group)
		require.NoError(b, err, "failed to create group %d", g)
		groups[g] = group
	}

	// 3. Create GroupPermissions
	for _, perm := range permissions {
		groupIdx := perm.ID % int64(numGroups)
		group := groups[groupIdx]

		gpUID := uuid.New().String()
		sql := `INSERT INTO group_permission (uid, group_id, permission_id) VALUES ($1, $2, $3)`
		_, err := db.Exec(ctx, sql, gpUID, group.ID, perm.ID)
		require.NoError(b, err, "failed to create group permission")
	}

	// 4. Create roles and RolePermissions
	roleRepo := repository.NewRoleRepository(db)
	rolesPerGroup := numRoles / numGroups
	remainingRoles := numRoles % numGroups
	createdRoles := make([]*model.Role, 0, numRoles)

	for g := 0; g < numGroups; g++ {
		rolesInThisGroup := rolesPerGroup
		if g < remainingRoles {
			rolesInThisGroup++
		}
		for r := 0; r < rolesInThisGroup; r++ {
			role := &model.Role{
				UID:         uuid.New().String(),
				GroupID:     groups[g].ID,
				Name:        fmt.Sprintf("benchmark-role-%d-%d", now, r),
				Description: fmt.Sprintf("Benchmark role %d in group %d", r, g),
			}
			err := roleRepo.Create(ctx, role)
			require.NoError(b, err, "failed to create role")
			createdRoles = append(createdRoles, role)

			// Get group_permissions
			groupPermSQL := `SELECT id FROM group_permission WHERE group_id = $1`
			rows, err := db.Query(ctx, groupPermSQL, groups[g].ID)
			if err != nil {
				b.Fatalf("failed to query group permissions: %v", err)
			}
			var groupPermIDs []int64
			for rows.Next() {
				var id int64
				if err := rows.Scan(&id); err != nil {
					b.Fatalf("failed to scan group_permission id: %v", err)
				}
				groupPermIDs = append(groupPermIDs, id)
			}
			rows.Close()

			// Insert role_permission
			rolePermSQL := `INSERT INTO role_permission (role_id, group_permission_id) VALUES ($1, $2)`
			for _, gpID := range groupPermIDs {
				_, err := db.Exec(ctx, rolePermSQL, role.ID, gpID)
				require.NoError(b, err, "failed to create role permission")
			}
		}
	}

	// 5. Create SubjectRole assignments
	subjectRepo := repository.NewSubjectRepository(db)
	for _, role := range createdRoles {
		sr := &model.SubjectRole{
			SubjectID:   subjectID,
			SubjectType: subjectType,
			RoleID:      role.ID,
		}
		err := subjectRepo.Create(ctx, sr)
		require.NoError(b, err, "failed to assign role to subject")
	}

	return subjectID, subjectType
}

// truncateDB truncates all tables cleanly
func truncateDB(t *testing.B, db *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	if err := truncateTestTables(ctx, db); err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}
}

// truncateTestTables uses the same approach as testutil/postgre.go
func truncateTestTables(ctx context.Context, db *pgxpool.Pool) error {
	// get all table names
	var tables []string
	err := db.QueryRow(ctx, `
		SELECT array_agg(tablename)
		FROM pg_tables
		WHERE schemaname = 'public' AND tablename != 'databasechangelog' AND tablename != 'databasechangeloglock'
	`).Scan(&tables)
	if err != nil {
		return fmt.Errorf("failed to get table names: %w", err)
	}

	if len(tables) == 0 {
		return fmt.Errorf("no tables found")
	}

	tx, err := db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	// truncate all tables in a single statement
	// using CASCADE to handle foreign key constraints automatically
	query := fmt.Sprintf(`TRUNCATE TABLE "%s"`, tables[0])
	for i := 1; i < len(tables); i++ {
		query += fmt.Sprintf(`, "%s"`, tables[i])
	}
	query += " RESTART IDENTITY CASCADE"

	if _, err := tx.Exec(ctx, query); err != nil {
		return fmt.Errorf("failed to truncate tables: %w", err)
	}

	return tx.Commit(ctx)
}

// getDB returns a database connection. Uses DATABASE_URL or constructs from env vars.
func getDB(t *testing.B) *pgxpool.Pool {
	t.Helper()

	// Try DATABASE_URL first
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL != "" {
		pool, err := pgxpool.New(context.Background(), dbURL)
		if err != nil {
			t.Fatalf("failed to create pool: %v", err)
		}
		t.Cleanup(func() { pool.Close() })
		return pool
	}

	// Construct from individual env vars
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     getEnv("DATABASE_HOST", "localhost"),
			Port:     getAtoi(getEnv("DATABASE_PORT", "5432"), 5432),
			User:     getEnv("DATABASE_USER", "postgres"),
			Password: getEnv("DATABASE_PASSWORD", "postgres"),
			Name:     getEnv("DATABASE_NAME", "service_db"),
			SslMode:  getEnv("DATABASE_SSL_MODE", "disable"),
		},
	}

	ctx := context.Background()
	pool, err := infra.NewPostgreConnection(ctx, &infra.PostgreConfig{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		Name:     cfg.Database.Name,
		SslMode:  cfg.Database.SslMode,
	})
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	t.Cleanup(func() { pool.Close() })

	// Truncate tables before each benchmark
	if err := truncateTestTables(ctx, pool); err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}

	return pool
}

// getEnv reads an environment variable with a default
func getEnv(key, defaultValue string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return defaultValue
}

// getAtoi converts string to int with default on error
func getAtoi(s string, def int) int {
	if i, err := strconv.Atoi(s); err == nil {
		return i
	}
	return def
}

// --- Custom repository provider for benchmarks ---

type customRepositoryProvider struct {
	db *pgxpool.Pool
}

func newCustomRepositoryProvider(db *pgxpool.Pool) portrepository.RepositoryProvider {
	return &customRepositoryProvider{db: db}
}

func (p *customRepositoryProvider) Subject() portrepository.SubjectRepository {
	return repository.NewSubjectRepository(p.db)
}
func (p *customRepositoryProvider) Role() portrepository.RoleRepository {
	return repository.NewRoleRepository(p.db)
}
func (p *customRepositoryProvider) Group() portrepository.GroupRepository {
	return repository.NewGroupRepository(p.db)
}
func (p *customRepositoryProvider) Permission() portrepository.PermissionRepository {
	return repository.NewPermissionRepository(p.db)
}

// --- Inline mocks for benchmark tests ---

type mockUnitOfWork struct{}

func (m *mockUnitOfWork) Do(ctx context.Context, fn func(portrepository.RepositoryProvider) error) error {
	return fn(nil) // Provider will be passed by the service
}

type mockObserver struct{}

func (m *mockObserver) OnSignal(ctx context.Context, sig signal.SignalType, data signal.SignalSubject, err error) {
	// No-op for benchmarks
}
