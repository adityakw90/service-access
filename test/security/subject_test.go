package security

import (
	"context"
	"testing"

	"github.com/adityakw90/service-access-proto/gen/go/common"
	"github.com/adityakw90/service-access-proto/gen/go/subject"
	"github.com/stretchr/testify/require"
)

// TestSubjectService_SQLInjection_List tests SQL injection via Subject List endpoint.
func TestSubjectService_SQLInjection_List(t *testing.T) {
	client := setupSecurityTest(t)
	defer client.Close()

	ctx := context.Background()

	for _, payload := range SQLInjectionPayloads() {
		t.Run(payload[:min(20, len(payload))], func(t *testing.T) {
			_, err := client.Client.SubjectClient.List(ctx, &subject.ListRequest{
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

// TestSubjectService_EmptyOrderBy tests empty OrderBy values via Subject List.
func TestSubjectService_EmptyOrderBy(t *testing.T) {
	client := setupSecurityTest(t)
	defer client.Close()

	ctx := context.Background()

	for _, payload := range EmptyOrderByPayloads() {
		t.Run("Empty_"+repr(payload), func(t *testing.T) {
			_, err := client.Client.SubjectClient.List(ctx, &subject.ListRequest{
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
