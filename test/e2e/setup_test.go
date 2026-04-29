package e2e

import (
	"context"
	"testing"

	testutil "github.com/adityakw90/service-access/test/util"
	"github.com/stretchr/testify/require"
)

// setupE2ETest creates and starts test infrastructure (services, gRPC server, client).
// Returns a cleanup function that should be called in defer.
func setupE2ETest(t *testing.T) (*testutil.TestGRPCClient, func()) {
	t.Helper()

	ctx := context.Background()

	// Setup test services
	testServices, err := testutil.SetupTestServices(t, ctx)
	require.NoError(t, err)

	// Start gRPC server
	grpcServer, err := testutil.NewTestGRPCServer(testServices)
	require.NoError(t, err)

	// Create gRPC clients
	grpcClient, err := testutil.NewTestGRPCClient(grpcServer.Addr())
	require.NoError(t, err)

	// Cleanup function
	cleanup := func() {
		grpcClient.Close()
		grpcServer.Stop()
	}

	return grpcClient, cleanup
}
