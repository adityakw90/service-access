package e2e

import (
	"context"
	"testing"
	"time"

	grouppb "github.com/adityakw90/service-access-proto/gen/go/group"
	permissionpb "github.com/adityakw90/service-access-proto/gen/go/permission"
	rolepb "github.com/adityakw90/service-access-proto/gen/go/role"
	subjectpb "github.com/adityakw90/service-access-proto/gen/go/subject"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TestE2E_SubjectAggregation tests the subject aggregation endpoints using table-driven tests
func TestE2E_SubjectAggregation(t *testing.T) {
	grpcClient, cleanup := setupE2ETest(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	setupCtx, setupCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer setupCancel()

	// Setup: Create permissions
	perm1, err := grpcClient.PermissionClient.Create(setupCtx, &permissionpb.CreateRequest{
		Resource:    "documents",
		Action:      "read",
		Description: "Read documents",
	})
	require.NoError(t, err)

	perm2, err := grpcClient.PermissionClient.Create(setupCtx, &permissionpb.CreateRequest{
		Resource:    "documents",
		Action:      "write",
		Description: "Write documents",
	})
	require.NoError(t, err)

	perm3, err := grpcClient.PermissionClient.Create(setupCtx, &permissionpb.CreateRequest{
		Resource:    "reports",
		Action:      "view",
		Description: "View reports",
	})
	require.NoError(t, err)

	// Setup: Create groups
	group1, err := grpcClient.GroupClient.Create(setupCtx, &grouppb.CreateRequest{
		Name:        "document-group",
		Description: "Group for document permissions",
	})
	require.NoError(t, err)

	_, err = grpcClient.GroupClient.UpdatePermission(setupCtx, &grouppb.UpdatePermissionRequest{
		GroupUid:       group1.Uid,
		PermissionUids: []string{perm1.Uid, perm2.Uid},
	})
	require.NoError(t, err)

	group2, err := grpcClient.GroupClient.Create(setupCtx, &grouppb.CreateRequest{
		Name:        "report-group",
		Description: "Group for report permissions",
	})
	require.NoError(t, err)

	_, err = grpcClient.GroupClient.UpdatePermission(setupCtx, &grouppb.UpdatePermissionRequest{
		GroupUid:       group2.Uid,
		PermissionUids: []string{perm3.Uid},
	})
	require.NoError(t, err)

	// Setup: Create roles
	role1, err := grpcClient.RoleClient.Create(setupCtx, &rolepb.CreateRequest{
		GroupUid:    group1.Uid,
		Name:        "reader",
		Description: "Can read documents",
	})
	require.NoError(t, err)

	role2, err := grpcClient.RoleClient.Create(setupCtx, &rolepb.CreateRequest{
		GroupUid:    group1.Uid,
		Name:        "writer",
		Description: "Can write documents",
	})
	require.NoError(t, err)

	role3, err := grpcClient.RoleClient.Create(setupCtx, &rolepb.CreateRequest{
		GroupUid:    group2.Uid,
		Name:        "viewer",
		Description: "Can view reports",
	})
	require.NoError(t, err)

	// Setup: Assign permissions to roles
	group1Perms, err := grpcClient.GroupClient.ListPermissions(setupCtx, &grouppb.ListPermissionsRequest{
		GroupUid: group1.Uid,
	})
	require.NoError(t, err)
	require.Len(t, group1Perms.Items, 2)

	var readGroupPermUID, writeGroupPermUID string
	for _, gp := range group1Perms.Items {
		if gp.PermissionUid == perm1.Uid {
			readGroupPermUID = gp.Uid
		}
		if gp.PermissionUid == perm2.Uid {
			writeGroupPermUID = gp.Uid
		}
	}
	require.NotEmpty(t, readGroupPermUID)
	require.NotEmpty(t, writeGroupPermUID)

	_, err = grpcClient.RoleClient.UpdatePermission(setupCtx, &rolepb.UpdatePermissionRequest{
		RoleUid:             role1.Uid,
		GroupPermissionUids: []string{readGroupPermUID},
	})
	require.NoError(t, err)

	_, err = grpcClient.RoleClient.UpdatePermission(setupCtx, &rolepb.UpdatePermissionRequest{
		RoleUid:             role2.Uid,
		GroupPermissionUids: []string{writeGroupPermUID},
	})
	require.NoError(t, err)

	group2Perms, err := grpcClient.GroupClient.ListPermissions(setupCtx, &grouppb.ListPermissionsRequest{
		GroupUid: group2.Uid,
	})
	require.NoError(t, err)
	require.Len(t, group2Perms.Items, 1)
	viewGroupPermUID := group2Perms.Items[0].Uid

	_, err = grpcClient.RoleClient.UpdatePermission(setupCtx, &rolepb.UpdatePermissionRequest{
		RoleUid:             role3.Uid,
		GroupPermissionUids: []string{viewGroupPermUID},
	})
	require.NoError(t, err)

	// Setup: Assign roles to subject
	subjectID := "user-aggregation-test"
	subjectType := "user"

	_, err = grpcClient.SubjectClient.AssignRole(setupCtx, &subjectpb.AssignRoleRequest{
		SubjectId:   subjectID,
		SubjectType: subjectType,
		RoleUid:     role1.Uid,
	})
	require.NoError(t, err)

	_, err = grpcClient.SubjectClient.AssignRole(setupCtx, &subjectpb.AssignRoleRequest{
		SubjectId:   subjectID,
		SubjectType: subjectType,
		RoleUid:     role2.Uid,
	})
	require.NoError(t, err)

	_, err = grpcClient.SubjectClient.AssignRole(setupCtx, &subjectpb.AssignRoleRequest{
		SubjectId:   subjectID,
		SubjectType: subjectType,
		RoleUid:     role3.Uid,
	})
	require.NoError(t, err)

	// Table-driven tests for aggregation endpoints
	tests := []struct {
		name     string
		endpoint string
		request  *subjectpb.GetSubjectRequest
		validate func(t *testing.T, resp interface{}, err error)
	}{
		{
			name:     "ListGroup returns all unique groups",
			endpoint: "ListGroup",
			request: &subjectpb.GetSubjectRequest{
				SubjectId:   subjectID,
				SubjectType: subjectType,
			},
			validate: func(t *testing.T, respInterface interface{}, err error) {
				require.NoError(t, err)
				resp, ok := respInterface.(*subjectpb.ListGroupResponse)
				require.True(t, ok, "Response should be ListGroupResponse")
				require.Len(t, resp.Groups, 2)
				require.Equal(t, int32(2), resp.Total)
				groupUIDs := make(map[string]bool)
				for _, g := range resp.Groups {
					groupUIDs[g.Uid] = true
				}
				require.True(t, groupUIDs[group1.Uid])
				require.True(t, groupUIDs[group2.Uid])
			},
		},
		{
			name:     "ListRole returns all roles",
			endpoint: "ListRole",
			request: &subjectpb.GetSubjectRequest{
				SubjectId:   subjectID,
				SubjectType: subjectType,
			},
			validate: func(t *testing.T, respInterface interface{}, err error) {
				require.NoError(t, err)
				resp, ok := respInterface.(*subjectpb.ListRoleResponse)
				require.True(t, ok, "Response should be ListRoleResponse")
				require.Len(t, resp.Roles, 3)
				require.Equal(t, int32(3), resp.Total)
				roleUIDs := make(map[string]bool)
				for _, r := range resp.Roles {
					roleUIDs[r.Uid] = true
				}
				require.True(t, roleUIDs[role1.Uid])
				require.True(t, roleUIDs[role2.Uid])
				require.True(t, roleUIDs[role3.Uid])
			},
		},
		{
			name:     "ListPermission returns deduplicated permissions",
			endpoint: "ListPermission",
			request: &subjectpb.GetSubjectRequest{
				SubjectId:   subjectID,
				SubjectType: subjectType,
			},
			validate: func(t *testing.T, respInterface interface{}, err error) {
				require.NoError(t, err)
				resp, ok := respInterface.(*subjectpb.ListPermissionResponse)
				require.True(t, ok, "Response should be ListPermissionResponse")
				require.Len(t, resp.Permissions, 3)
				require.Equal(t, int32(3), resp.Total)
				permUIDs := make(map[string]bool)
				for _, p := range resp.Permissions {
					permUIDs[p.Uid] = true
				}
				require.True(t, permUIDs[perm1.Uid])
				require.True(t, permUIDs[perm2.Uid])
				require.True(t, permUIDs[perm3.Uid])
			},
		},
		{
			name:     "Get returns full profile",
			endpoint: "Get",
			request: &subjectpb.GetSubjectRequest{
				SubjectId:   subjectID,
				SubjectType: subjectType,
			},
			validate: func(t *testing.T, respInterface interface{}, err error) {
				require.NoError(t, err)
				resp, ok := respInterface.(*subjectpb.GetSubjectResponse)
				require.True(t, ok, "Response should be GetSubjectResponse")
				require.Len(t, resp.Groups, 2)
				require.Equal(t, int32(2), resp.TotalGroup)
				require.Len(t, resp.Roles, 3)
				require.Equal(t, int32(3), resp.TotalRole)
				require.Len(t, resp.Permissions, 3)
				require.Equal(t, int32(3), resp.TotalPermission)
			},
		},
		{
			name:     "ListGroup for non-existent subject returns empty",
			endpoint: "ListGroup",
			request: &subjectpb.GetSubjectRequest{
				SubjectId:   "non-existent",
				SubjectType: "user",
			},
			validate: func(t *testing.T, respInterface interface{}, err error) {
				require.NoError(t, err)
				resp, ok := respInterface.(*subjectpb.ListGroupResponse)
				require.True(t, ok)
				require.Len(t, resp.Groups, 0)
				require.Equal(t, int32(0), resp.Total)
			},
		},
		{
			name:     "ListPermission for non-existent subject returns empty",
			endpoint: "ListPermission",
			request: &subjectpb.GetSubjectRequest{
				SubjectId:   "non-existent",
				SubjectType: "user",
			},
			validate: func(t *testing.T, respInterface interface{}, err error) {
				require.NoError(t, err)
				resp, ok := respInterface.(*subjectpb.ListPermissionResponse)
				require.True(t, ok)
				require.Len(t, resp.Permissions, 0)
				require.Equal(t, int32(0), resp.Total)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp interface{}
			var err error

			switch tt.endpoint {
			case "ListGroup":
				resp, err = grpcClient.SubjectClient.ListGroup(ctx, tt.request)
			case "ListRole":
				resp, err = grpcClient.SubjectClient.ListRole(ctx, tt.request)
			case "ListPermission":
				resp, err = grpcClient.SubjectClient.ListPermission(ctx, tt.request)
			case "Get":
				resp, err = grpcClient.SubjectClient.Get(ctx, tt.request)
			}

			tt.validate(t, resp, err)
		})
	}

	// Separate table for validation errors
	t.Run("Validation Errors", func(t *testing.T) {
		validationTests := []struct {
			name        string
			endpoint    string
			request     *subjectpb.GetSubjectRequest
			wantCode    codes.Code
			wantMessage string
		}{
			{
				name:     "ListGroup missing subject_id",
				endpoint: "ListGroup",
				request:  &subjectpb.GetSubjectRequest{SubjectType: "user"},
				wantCode: codes.InvalidArgument,
			},
			{
				name:     "ListRole missing subject_type",
				endpoint: "ListRole",
				request:  &subjectpb.GetSubjectRequest{SubjectId: "user-123"},
				wantCode: codes.InvalidArgument,
			},
			{
				name:        "ListPermission empty request",
				endpoint:    "ListPermission",
				request:     &subjectpb.GetSubjectRequest{},
				wantCode:    codes.InvalidArgument,
				wantMessage: "SubjectID is required",
			},
			{
				name:        "Get missing SubjectID",
				endpoint:    "Get",
				request:     &subjectpb.GetSubjectRequest{SubjectType: "user"},
				wantCode:    codes.InvalidArgument,
				wantMessage: "SubjectID is required",
			},
		}

		for _, tt := range validationTests {
			t.Run(tt.name, func(t *testing.T) {
				var err error
				switch tt.endpoint {
				case "ListGroup":
					_, err = grpcClient.SubjectClient.ListGroup(ctx, tt.request)
				case "ListRole":
					_, err = grpcClient.SubjectClient.ListRole(ctx, tt.request)
				case "ListPermission":
					_, err = grpcClient.SubjectClient.ListPermission(ctx, tt.request)
				case "Get":
					_, err = grpcClient.SubjectClient.Get(ctx, tt.request)
				}

				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, tt.wantCode, st.Code())
				if tt.wantMessage != "" {
					require.Contains(t, st.Message(), tt.wantMessage)
				}
			})
		}
	})

	// Deduplication test
	t.Run("Deduplication of permissions", func(t *testing.T) {
		// Create a new group with duplicate permission (perm1)
		dupGroup, err := grpcClient.GroupClient.Create(setupCtx, &grouppb.CreateRequest{
			Name:        "duplicate-group",
			Description: "Group with duplicate permission",
		})
		require.NoError(t, err)

		_, err = grpcClient.GroupClient.UpdatePermission(setupCtx, &grouppb.UpdatePermissionRequest{
			GroupUid:       dupGroup.Uid,
			PermissionUids: []string{perm1.Uid},
		})
		require.NoError(t, err)

		dupRole, err := grpcClient.RoleClient.Create(setupCtx, &rolepb.CreateRequest{
			GroupUid:    dupGroup.Uid,
			Name:        "dup-role",
			Description: "Role with duplicate permission",
		})
		require.NoError(t, err)

		dupGroupPerms, err := grpcClient.GroupClient.ListPermissions(setupCtx, &grouppb.ListPermissionsRequest{
			GroupUid: dupGroup.Uid,
		})
		require.NoError(t, err)
		require.Len(t, dupGroupPerms.Items, 1)

		_, err = grpcClient.RoleClient.UpdatePermission(setupCtx, &rolepb.UpdatePermissionRequest{
			RoleUid:             dupRole.Uid,
			GroupPermissionUids: []string{dupGroupPerms.Items[0].Uid},
		})
		require.NoError(t, err)

		_, err = grpcClient.SubjectClient.AssignRole(setupCtx, &subjectpb.AssignRoleRequest{
			SubjectId:   subjectID,
			SubjectType: subjectType,
			RoleUid:     dupRole.Uid,
		})
		require.NoError(t, err)

		// Subject now has perm1 from both reader and dup-role, but should see it only once
		resp, err := grpcClient.SubjectClient.ListPermission(ctx, &subjectpb.GetSubjectRequest{
			SubjectId:   subjectID,
			SubjectType: subjectType,
		})
		require.NoError(t, err)
		require.Len(t, resp.Permissions, 3, "Should have 3 unique permissions (deduplicated)")

		permCount := 0
		for _, p := range resp.Permissions {
			if p.Uid == perm1.Uid {
				permCount++
			}
		}
		require.Equal(t, 1, permCount, "Permission should appear exactly once")
	})
}
