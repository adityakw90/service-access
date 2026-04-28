package e2e

import (
	"context"
	"testing"
	"time"

	accesspb "github.com/adityakw90/service-access-proto/gen/go/access"
	grouppb "github.com/adityakw90/service-access-proto/gen/go/group"
	permissionpb "github.com/adityakw90/service-access-proto/gen/go/permission"
	rolepb "github.com/adityakw90/service-access-proto/gen/go/role"
	subjectpb "github.com/adityakw90/service-access-proto/gen/go/subject"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TestE2E_AccessService_CheckAccess is a table-driven test for access control checks.
func TestE2E_AccessService_CheckAccess(t *testing.T) {
	grpcClient, cleanup := setupE2ETest(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Setup test data: create permissions, group, roles, and subject assignments
	setupCtx, setupCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer setupCancel()

	// Create permissions for testing
	readInvoicesPerm, err := grpcClient.PermissionClient.Create(setupCtx, &permissionpb.CreateRequest{
		Resource:    "invoices",
		Action:      "read",
		Description: "Permission to read invoices",
	})
	require.NoError(t, err)

	writeInvoicesPerm, err := grpcClient.PermissionClient.Create(setupCtx, &permissionpb.CreateRequest{
		Resource:    "invoices",
		Action:      "write",
		Description: "Permission to write invoices",
	})
	require.NoError(t, err)

	_, err = grpcClient.PermissionClient.Create(setupCtx, &permissionpb.CreateRequest{
		Resource:    "users",
		Action:      "delete",
		Description: "Permission to delete users",
	})
	require.NoError(t, err)

	// Create a group and assign permissions
	accountingGroup, err := grpcClient.GroupClient.Create(setupCtx, &grouppb.CreateRequest{
		Name:        "accounting",
		Description: "Accounting department group",
	})
	require.NoError(t, err)

	_, err = grpcClient.GroupClient.UpdatePermission(setupCtx, &grouppb.UpdatePermissionRequest{
		GroupUid:       accountingGroup.Uid,
		PermissionUids: []string{readInvoicesPerm.Uid, writeInvoicesPerm.Uid},
	})
	require.NoError(t, err)

	// Create roles within the group
	accountantRole, err := grpcClient.RoleClient.Create(setupCtx, &rolepb.CreateRequest{
		GroupUid:    accountingGroup.Uid,
		Name:        "accountant",
		Description: "Accountant role",
	})
	require.NoError(t, err)

	// Assign read invoices permission to accountant role
	readInvoicesGroupPerm, err := grpcClient.GroupClient.ListPermissions(setupCtx, &grouppb.ListPermissionsRequest{
		GroupUid: accountingGroup.Uid,
	})
	require.NoError(t, err)
	var readInvoicesGroupPermUID string
	for _, gp := range readInvoicesGroupPerm.Items {
		if gp.PermissionUid == readInvoicesPerm.Uid {
			readInvoicesGroupPermUID = gp.Uid
			break
		}
	}
	require.NotEmpty(t, readInvoicesGroupPermUID)

	_, err = grpcClient.RoleClient.UpdatePermission(setupCtx, &rolepb.UpdatePermissionRequest{
		RoleUid:             accountantRole.Uid,
		GroupPermissionUids: []string{readInvoicesGroupPermUID},
	})
	require.NoError(t, err)

	// Assign role to subject
	_, err = grpcClient.SubjectClient.AssignRole(setupCtx, &subjectpb.AssignRoleRequest{
		SubjectId:   "user-123",
		SubjectType: "user",
		RoleUid:     accountantRole.Uid,
	})
	require.NoError(t, err)

	// Define test cases
	tests := []struct {
		name        string
		subjectID   string
		subjectType string
		resource    string
		action      string
		wantAllowed bool
		wantReason  string
	}{
		{
			name:        "Allowed: user with read permission can read invoices",
			subjectID:   "user-123",
			subjectType: "user",
			resource:    "invoices",
			action:      "read",
			wantAllowed: true,
			wantReason:  "permission granted",
		},
		{
			name:        "Denied: user without write permission cannot write invoices",
			subjectID:   "user-123",
			subjectType: "user",
			resource:    "invoices",
			action:      "write",
			wantAllowed: false,
			wantReason:  "no matching permission found",
		},
		{
			name:        "Denied: user without delete users permission",
			subjectID:   "user-123",
			subjectType: "user",
			resource:    "users",
			action:      "delete",
			wantAllowed: false,
			wantReason:  "no matching permission found",
		},
		{
			name:        "Denied: unknown subject",
			subjectID:   "unknown-user",
			subjectType: "user",
			resource:    "invoices",
			action:      "read",
			wantAllowed: false,
			wantReason:  "no matching permission found",
		},
		{
			name:        "Denied: different resource",
			subjectID:   "user-123",
			subjectType: "user",
			resource:    "payments",
			action:      "read",
			wantAllowed: false,
			wantReason:  "no matching permission found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := grpcClient.AccessClient.CheckAccess(ctx, &accesspb.CheckAccessRequest{
				SubjectId:   tt.subjectID,
				SubjectType: tt.subjectType,
				Resource:    tt.resource,
				Action:      tt.action,
			})
			require.NoError(t, err)

			if resp.Allowed != tt.wantAllowed {
				t.Errorf("CheckAccess() allowed = %v, want %v", resp.Allowed, tt.wantAllowed)
			}

			// For allowed requests, reason should contain "permission granted"
			// For denied requests, reason should contain "no matching permission found"
			if tt.wantAllowed {
				require.Contains(t, resp.Reason, tt.wantReason, "reason should indicate permission granted")
			} else {
				require.Contains(t, resp.Reason, tt.wantReason, "reason should indicate no permission found")
			}
		})
	}
}

// TestE2E_AccessService_CheckAccess_WithMultipleRoles tests access with multiple role assignments.
func TestE2E_AccessService_CheckAccess_WithMultipleRoles(t *testing.T) {
	grpcClient, cleanup := setupE2ETest(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Setup test data
	setupCtx, setupCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer setupCancel()

	// Create permissions
	readReportsPerm, err := grpcClient.PermissionClient.Create(setupCtx, &permissionpb.CreateRequest{
		Resource:    "reports",
		Action:      "read",
		Description: "Read reports",
	})
	require.NoError(t, err)

	writeReportsPerm, err := grpcClient.PermissionClient.Create(setupCtx, &permissionpb.CreateRequest{
		Resource:    "reports",
		Action:      "write",
		Description: "Write reports",
	})
	require.NoError(t, err)

	// Create group with permissions
	reportsGroup, err := grpcClient.GroupClient.Create(setupCtx, &grouppb.CreateRequest{
		Name:        "reports",
		Description: "Reports management",
	})
	require.NoError(t, err)

	_, err = grpcClient.GroupClient.UpdatePermission(setupCtx, &grouppb.UpdatePermissionRequest{
		GroupUid:       reportsGroup.Uid,
		PermissionUids: []string{readReportsPerm.Uid, writeReportsPerm.Uid},
	})
	require.NoError(t, err)

	// Create two roles: one with read, one with write
	readerRole, err := grpcClient.RoleClient.Create(setupCtx, &rolepb.CreateRequest{
		GroupUid:    reportsGroup.Uid,
		Name:        "reader",
		Description: "Reader role",
	})
	require.NoError(t, err)

	writerRole, err := grpcClient.RoleClient.Create(setupCtx, &rolepb.CreateRequest{
		GroupUid:    reportsGroup.Uid,
		Name:        "writer",
		Description: "Writer role",
	})
	require.NoError(t, err)

	// Get group permission UIDs
	groupPerms, err := grpcClient.GroupClient.ListPermissions(setupCtx, &grouppb.ListPermissionsRequest{
		GroupUid: reportsGroup.Uid,
	})
	require.NoError(t, err)

	var readGroupPermUID, writeGroupPermUID string
	for _, gp := range groupPerms.Items {
		if gp.PermissionUid == readReportsPerm.Uid {
			readGroupPermUID = gp.Uid
		}
		if gp.PermissionUid == writeReportsPerm.Uid {
			writeGroupPermUID = gp.Uid
		}
	}
	require.NotEmpty(t, readGroupPermUID)
	require.NotEmpty(t, writeGroupPermUID)

	// Assign permissions to roles
	_, err = grpcClient.RoleClient.UpdatePermission(setupCtx, &rolepb.UpdatePermissionRequest{
		RoleUid:             readerRole.Uid,
		GroupPermissionUids: []string{readGroupPermUID},
	})
	require.NoError(t, err)

	_, err = grpcClient.RoleClient.UpdatePermission(setupCtx, &rolepb.UpdatePermissionRequest{
		RoleUid:             writerRole.Uid,
		GroupPermissionUids: []string{writeGroupPermUID},
	})
	require.NoError(t, err)

	// Assign both roles to a subject
	subjectID := "multi-role-user"
	_, err = grpcClient.SubjectClient.AssignRole(setupCtx, &subjectpb.AssignRoleRequest{
		SubjectId:   subjectID,
		SubjectType: "user",
		RoleUid:     readerRole.Uid,
	})
	require.NoError(t, err)

	_, err = grpcClient.SubjectClient.AssignRole(setupCtx, &subjectpb.AssignRoleRequest{
		SubjectId:   subjectID,
		SubjectType: "user",
		RoleUid:     writerRole.Uid,
	})
	require.NoError(t, err)

	// Test cases
	tests := []struct {
		name        string
		resource    string
		action      string
		wantAllowed bool
	}{
		{
			name:        "Can read reports (from reader role)",
			resource:    "reports",
			action:      "read",
			wantAllowed: true,
		},
		{
			name:        "Can write reports (from writer role)",
			resource:    "reports",
			action:      "write",
			wantAllowed: true,
		},
		{
			name:        "Cannot delete reports",
			resource:    "reports",
			action:      "delete",
			wantAllowed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := grpcClient.AccessClient.CheckAccess(ctx, &accesspb.CheckAccessRequest{
				SubjectId:   subjectID,
				SubjectType: "user",
				Resource:    tt.resource,
				Action:      tt.action,
			})
			require.NoError(t, err)

			if resp.Allowed != tt.wantAllowed {
				t.Errorf("CheckAccess() allowed = %v, want %v, reason: %s", resp.Allowed, tt.wantAllowed, resp.Reason)
			}
		})
	}
}

// TestE2E_AccessService_CheckAccess_WithRoleRevocation tests access after role revocation.
func TestE2E_AccessService_CheckAccess_WithRoleRevocation(t *testing.T) {
	grpcClient, cleanup := setupE2ETest(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Setup test data
	setupCtx, setupCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer setupCancel()

	// Create permission
	perm, err := grpcClient.PermissionClient.Create(setupCtx, &permissionpb.CreateRequest{
		Resource:    "documents",
		Action:      "edit",
		Description: "Edit documents",
	})
	require.NoError(t, err)

	// Create group
	group, err := grpcClient.GroupClient.Create(setupCtx, &grouppb.CreateRequest{
		Name:        "documents",
		Description: "Document management",
	})
	require.NoError(t, err)

	_, err = grpcClient.GroupClient.UpdatePermission(setupCtx, &grouppb.UpdatePermissionRequest{
		GroupUid:       group.Uid,
		PermissionUids: []string{perm.Uid},
	})
	require.NoError(t, err)

	// Create role
	role, err := grpcClient.RoleClient.Create(setupCtx, &rolepb.CreateRequest{
		GroupUid:    group.Uid,
		Name:        "editor",
		Description: "Editor role",
	})
	require.NoError(t, err)

	// Get group permission UID
	groupPerms, err := grpcClient.GroupClient.ListPermissions(setupCtx, &grouppb.ListPermissionsRequest{
		GroupUid: group.Uid,
	})
	require.NoError(t, err)
	require.Len(t, groupPerms.Items, 1)
	groupPermUID := groupPerms.Items[0].Uid

	_, err = grpcClient.RoleClient.UpdatePermission(setupCtx, &rolepb.UpdatePermissionRequest{
		RoleUid:             role.Uid,
		GroupPermissionUids: []string{groupPermUID},
	})
	require.NoError(t, err)

	// Assign role to subject
	subjectID := "editor-user"
	_, err = grpcClient.SubjectClient.AssignRole(setupCtx, &subjectpb.AssignRoleRequest{
		SubjectId:   subjectID,
		SubjectType: "user",
		RoleUid:     role.Uid,
	})
	require.NoError(t, err)

	// Initially, user should have access
	resp, err := grpcClient.AccessClient.CheckAccess(ctx, &accesspb.CheckAccessRequest{
		SubjectId:   subjectID,
		SubjectType: "user",
		Resource:    "documents",
		Action:      "edit",
	})
	require.NoError(t, err)
	require.True(t, resp.Allowed, "User should initially have access")

	// Revoke the role
	_, err = grpcClient.SubjectClient.RevokeRole(setupCtx, &subjectpb.RevokeRoleRequest{
		SubjectId:   subjectID,
		SubjectType: "user",
		RoleUid:     role.Uid,
	})
	require.NoError(t, err)

	// After revocation, user should not have access
	resp, err = grpcClient.AccessClient.CheckAccess(ctx, &accesspb.CheckAccessRequest{
		SubjectId:   subjectID,
		SubjectType: "user",
		Resource:    "documents",
		Action:      "edit",
	})
	require.NoError(t, err)
	require.False(t, resp.Allowed, "User should not have access after role revocation")
	require.Contains(t, resp.Reason, "no matching permission found")
}

// TestE2E_AccessService_CheckAccess_InvalidRequests tests access with invalid requests.
func TestE2E_AccessService_CheckAccess_InvalidRequests(t *testing.T) {
	grpcClient, cleanup := setupE2ETest(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tests := []struct {
		name        string
		request     *accesspb.CheckAccessRequest
		wantCode    codes.Code
		wantMessage string
	}{
		{
			name: "Missing subject_id",
			request: &accesspb.CheckAccessRequest{
				SubjectType: "user",
				Resource:    "test",
				Action:      "read",
			},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "Missing subject_type",
			request: &accesspb.CheckAccessRequest{
				SubjectId: "user-123",
				Resource:  "test",
				Action:    "read",
			},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "Missing resource",
			request: &accesspb.CheckAccessRequest{
				SubjectId:   "user-123",
				SubjectType: "user",
				Action:      "read",
			},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "Missing action",
			request: &accesspb.CheckAccessRequest{
				SubjectId:   "user-123",
				SubjectType: "user",
				Resource:    "test",
			},
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "Empty request",
			request:  &accesspb.CheckAccessRequest{},
			wantCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := grpcClient.AccessClient.CheckAccess(ctx, tt.request)
			require.Error(t, err)

			st, ok := status.FromError(err)
			require.True(t, ok, "Error should be a gRPC status error")
			require.Equal(t, tt.wantCode, st.Code(), "Expected status code %v, got %v", tt.wantCode, st.Code())
		})
	}
}
