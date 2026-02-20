package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/adityakw90/service-access/internal/adapter/repository"
	"github.com/adityakw90/service-access/internal/adapter/security"
	"github.com/adityakw90/service-access/internal/core/domain/event"
	"github.com/adityakw90/service-access/internal/core/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()

	// Load configuration
	dbHost := getEnv("DATABASE_HOST", "localhost")
	dbPort := getEnv("DATABASE_PORT", "5432")
	dbName := getEnv("DATABASE_NAME", "vexa")
	dbUser := getEnv("DATABASE_USER", "vexa")
	dbPassword := getEnv("DATABASE_PASSWORD", "vexa")

	// Database connection
	connStr := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
		dbHost, dbPort, dbName, dbUser, dbPassword)

	db, err := setupDatabase(ctx, connStr)
	if err != nil {
		log.Fatalf("failed to setup database: %v", err)
	}
	defer db.Close()

	// Unit of Work for transactions
	uow := repository.NewUnitOfWork(db)

	// RepositoryProvider for reads
	repos := repository.NewRepositoryProvider(db)

	// Event Publisher - using noop publisher for now
	publisher := &noopPublisher{}

	// UID Generator
	uidGenerator := security.NewUIDGenerator()

	// Services
	groupService := service.NewGroupService(uow, repos, publisher, uidGenerator)
	permissionService := service.NewPermissionService(uow, repos, publisher, uidGenerator)
	roleService := service.NewRoleService(uow, repos, publisher, uidGenerator)
	subjectService := service.NewSubjectService(uow, repos, publisher)
	accessService := service.NewAccessService(repos)

	// TODO: Register with gRPC server when handlers are implemented
	log.Printf("Services initialized:")
	log.Printf("  GroupService: %T", groupService)
	log.Printf("  PermissionService: %T", permissionService)
	log.Printf("  RoleService: %T", roleService)
	log.Printf("  SubjectService: %T", subjectService)
	log.Printf("  AccessService: %T", accessService)

	// Keep running
	select {}
}

func setupDatabase(ctx context.Context, connStr string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %v", err)
	}

	return pool, nil
}

// noopPublisher implements portEvent.EventPublisher interface for development/testing
type noopPublisher struct{}

func (n *noopPublisher) Publish(ctx context.Context, eventType event.EventType, eventData any) error {
	log.Printf("Published event: %s (noop)", eventType)
	return nil
}

func (n *noopPublisher) Close() error {
	log.Printf("Closing noopPublisher")
	return nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
