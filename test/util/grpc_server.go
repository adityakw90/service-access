package testutil

import (
	"fmt"
	"time"

	grpcAdapter "github.com/adityakw90/service-access/internal/adapter/api/grpc"
)

// NewTestGRPCServer creates and starts a test gRPC server.
// Uses the provided TestServices for all dependencies.
func NewTestGRPCServer(testServices *TestServices) (*grpcAdapter.Server, error) {
	// Create server with test services
	server := grpcAdapter.NewServer(
		testServices.PermissionService,
		testServices.RoleService,
		testServices.GroupService,
		testServices.AccessService,
		testServices.SubjectService,
		testServices.Monitoring,
	)

	// Start server in background
	go func() {
		server.Start("127.0.0.1:0")
	}()

	// Wait for server to be ready (with timeout)
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if addr := server.Addr(); addr != "" {
			return server, nil
		}
		time.Sleep(10 * time.Millisecond)
	}

	return nil, fmt.Errorf("server failed to start within timeout")
}
