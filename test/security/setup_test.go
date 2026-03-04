package security

import (
	"context"
	"testing"

	testutil "github.com/adityakw90/service-access/test/util"

	"github.com/adityakw90/service-access-proto/gen/go/group"
	"github.com/adityakw90/service-access-proto/gen/go/permission"
	"github.com/adityakw90/service-access-proto/gen/go/role"
	"github.com/adityakw90/service-access-proto/gen/go/subject"

	"github.com/stretchr/testify/require"
)

// ==============================================================================
// gRPC Test Setup (follows E2E pattern)
// ==============================================================================

// setupSecurityTest creates test infrastructure for gRPC-level security testing.
// Returns a GRPCTestClient with test data and a cleanup function.
func setupSecurityTest(t *testing.T) *GRPCTestClient {
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

	// Create test data via gRPC
	groupResp, err := grpcClient.GroupClient.Create(ctx, &group.CreateRequest{
		Name:        "security-test-group",
		Description: "Group for security testing",
	})
	require.NoError(t, err)

	permResp, err := grpcClient.PermissionClient.Create(ctx, &permission.CreateRequest{
		Resource:    "security_test_resource",
		Action:      "test_action",
		Description: "Permission for security testing",
	})
	require.NoError(t, err)

	_, err = grpcClient.GroupClient.AssignPermission(ctx, &group.AssignPermissionRequest{
		GroupUid:      groupResp.Uid,
		PermissionUid: permResp.Uid,
	})
	require.NoError(t, err)

	roleResp, err := grpcClient.RoleClient.Create(ctx, &role.CreateRequest{
		GroupUid:    groupResp.Uid,
		Name:        "security-test-role",
		Description: "Role for security testing",
	})
	require.NoError(t, err)

	subjectID := "security-test-subject"
	subjectType := "user"
	_, err = grpcClient.SubjectClient.AssignRole(ctx, &subject.AssignRoleRequest{
		SubjectId:   subjectID,
		SubjectType: subjectType,
		RoleUid:     roleResp.Uid,
	})
	require.NoError(t, err)

	// Cleanup function
	cleanup := func() {
		ctx := context.Background()
		// Cleanup test data
		_, _ = grpcClient.SubjectClient.RevokeRole(ctx, &subject.RevokeRoleRequest{
			SubjectId:   subjectID,
			SubjectType: subjectType,
			RoleUid:     roleResp.Uid,
		})
		_, _ = grpcClient.RoleClient.Delete(ctx, &role.DeleteRequest{Uid: roleResp.Uid})
		_, _ = grpcClient.GroupClient.RevokePermission(ctx, &group.RevokePermissionRequest{
			GroupUid:      groupResp.Uid,
			PermissionUid: permResp.Uid,
		})
		_, _ = grpcClient.PermissionClient.Delete(ctx, &permission.DeleteRequest{Uid: permResp.Uid})
		_, _ = grpcClient.GroupClient.Delete(ctx, &group.DeleteRequest{Uid: groupResp.Uid})

		grpcClient.Close()
		grpcServer.Stop()
	}

	return &GRPCTestClient{
		Client:      grpcClient,
		GroupUID:    groupResp.Uid,
		PermUID:     permResp.Uid,
		RoleUID:     roleResp.Uid,
		SubjectID:   subjectID,
		SubjectType: subjectType,
		cleanup:     cleanup,
	}
}

// GRPCTestClient wraps the gRPC client for security testing.
type GRPCTestClient struct {
	Client      *testutil.TestGRPCClient
	GroupUID    string
	PermUID     string
	RoleUID     string
	SubjectID   string
	SubjectType string
	cleanup     func()
}

// Close cleans up the gRPC test client and server.
func (c *GRPCTestClient) Close() {
	if c.cleanup != nil {
		c.cleanup()
	}
}

// ==============================================================================
// Test Payloads
// ==============================================================================

// SQLInjectionPayloads returns common SQL injection payloads for testing.
func SQLInjectionPayloads() []string {
	return []string{
		"created_at; DROP TABLE permission; --",
		"created_at; DELETE FROM users WHERE 1=1; --",
		"created_at' OR '1'='1",
		"created_at' UNION SELECT * FROM users --",
		"created_at; INSERT INTO users (admin) VALUES (1); --",
		"created_at; UPDATE users SET admin=1 WHERE 1=1; --",
		"created_at; SHUTDOWN; --",
		"created_at'; EXEC xp_cmdshell('format c:'); --",
		"(SELECT CASE WHEN (1=1) THEN created_at ELSE id END)",
		"created_at,(SELECT password FROM users LIMIT 1)",
		"; DROP TABLE test; --",
		"nonexistent_column",
		"''; DROP TABLE test; --",
		"created_at; WAITFOR DELAY '00:00:10'; --",
	}
}

// EmptyOrderByPayloads returns empty and whitespace-only OrderBy values for edge case testing.
func EmptyOrderByPayloads() []string {
	return []string{
		"",           // Empty string
		"  ",         // Only spaces
		"\t",         // Only tab
		"\n",         // Only newline
		"  \t\n  ",   // Mixed whitespace
	}
}

// ==============================================================================
// Utilities
// ==============================================================================

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// repr returns a string representation of the payload for test naming.
func repr(s string) string {
	if s == "" {
		return "empty"
	}
	switch s {
	case "  ":
		return "spaces"
	case "\t":
		return "tab"
	case "\n":
		return "newline"
	case "  \t\n  ":
		return "mixed_whitespace"
	default:
		return "other"
	}
}
