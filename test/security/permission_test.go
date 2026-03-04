package security

import (
	"context"
	"testing"

	"github.com/adityakw90/service-access-proto/gen/go/common"
	"github.com/adityakw90/service-access-proto/gen/go/permission"
	"github.com/stretchr/testify/require"
)

// TestPermissionService_SQLInjection_List tests SQL injection via Permission List endpoint.
func TestPermissionService_SQLInjection_List(t *testing.T) {
	client := setupSecurityTest(t)
	defer client.Close()

	ctx := context.Background()

	for _, payload := range SQLInjectionPayloads() {
		t.Run(payload[:min(20, len(payload))], func(t *testing.T) {
			_, err := client.Client.PermissionClient.List(ctx, &permission.ListRequest{
				Pagination: &common.Pagination{
					Page:    1,
					Limit:   10,
					Sort:    "asc",
					OrderBy: payload,
				},
			})

			if err != nil {
				require.NotContains(t, err.Error(), "syntax error", "SQL injection detected!")
				require.NotContains(t, err.Error(), "DROP", "SQL injection detected!")
				require.NotContains(t, err.Error(), "DELETE", "SQL injection detected!")
				require.NotContains(t, err.Error(), "UNION", "SQL injection detected!")
			}
		})
	}
}

// TestPermissionService_EmptyOrderBy tests empty OrderBy values via Permission List.
func TestPermissionService_EmptyOrderBy(t *testing.T) {
	client := setupSecurityTest(t)
	defer client.Close()

	ctx := context.Background()

	for _, payload := range EmptyOrderByPayloads() {
		t.Run("Empty_"+repr(payload), func(t *testing.T) {
			_, err := client.Client.PermissionClient.List(ctx, &permission.ListRequest{
				Pagination: &common.Pagination{
					Page:    1,
					Limit:   10,
					Sort:    "asc",
					OrderBy: payload,
				},
			})

			if err != nil {
				require.NotContains(t, err.Error(), "syntax error", "Empty OrderBy caused SQL error")
			}
		})
	}
}
