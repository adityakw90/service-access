package testutil

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	accessgrpc "github.com/adityakw90/service-access-proto/gen/go/access"
	groupgrpc "github.com/adityakw90/service-access-proto/gen/go/group"
	permissiongrpc "github.com/adityakw90/service-access-proto/gen/go/permission"
	rolegrpc "github.com/adityakw90/service-access-proto/gen/go/role"
	subjectgrpc "github.com/adityakw90/service-access-proto/gen/go/subject"
)

// TestGRPCClient holds gRPC client connections for all services.
type TestGRPCClient struct {
	conn             *grpc.ClientConn
	AccessClient     accessgrpc.AccessControlServiceClient
	PermissionClient permissiongrpc.PermissionServiceClient
	GroupClient      groupgrpc.GroupServiceClient
	RoleClient       rolegrpc.RoleServiceClient
	SubjectClient    subjectgrpc.SubjectServiceClient
}

// NewTestGRPCClient creates gRPC clients connected to the test server.
func NewTestGRPCClient(serverAddr string) (*TestGRPCClient, error) {
	// Create connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	return &TestGRPCClient{
		conn:             conn,
		AccessClient:     accessgrpc.NewAccessControlServiceClient(conn),
		PermissionClient: permissiongrpc.NewPermissionServiceClient(conn),
		GroupClient:      groupgrpc.NewGroupServiceClient(conn),
		RoleClient:       rolegrpc.NewRoleServiceClient(conn),
		SubjectClient:    subjectgrpc.NewSubjectServiceClient(conn),
	}, nil
}

// Close closes the client connection.
func (c *TestGRPCClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}
