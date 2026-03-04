package security

import (
	"context"
	"testing"

	"github.com/adityakw90/service-access-proto/gen/go/common"
	"github.com/adityakw90/service-access-proto/gen/go/group"
	"github.com/stretchr/testify/require"
)

// TestGroupService_SQLInjection_List tests SQL injection via Group List endpoint.
func TestGroupService_SQLInjection_List(t *testing.T) {
	client := setupSecurityTest(t)
	defer client.Close()

	ctx := context.Background()

	for _, payload := range SQLInjectionPayloads() {
		t.Run(payload[:min(20, len(payload))], func(t *testing.T) {
			_, err := client.Client.GroupClient.List(ctx, &group.ListRequest{
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

// TestGroupService_EmptyOrderBy tests empty OrderBy values via Group List.
func TestGroupService_EmptyOrderBy(t *testing.T) {
	client := setupSecurityTest(t)
	defer client.Close()

	ctx := context.Background()

	for _, payload := range EmptyOrderByPayloads() {
		t.Run("Empty_"+repr(payload), func(t *testing.T) {
			_, err := client.Client.GroupClient.List(ctx, &group.ListRequest{
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
