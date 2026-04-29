package resolver

import (
	"context"
	"testing"

	"github.com/adityakw90/service-access/internal/core/domain/param"
	"github.com/adityakw90/service-access/internal/infra"
	"github.com/alicebob/miniredis/v2"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPermissionResolverIDsByResourceActions(t *testing.T) {
	tests := []struct {
		name            string
		resourceActions []param.PermissionMapResourceAction
		setup           func(pgxmock.PgxPoolIface)
		want            map[param.PermissionMapResourceAction]int64
		wantErr         bool
		errMsg          string
	}{
		{
			name: "Happy Path - single resource action",
			resourceActions: []param.PermissionMapResourceAction{
				{Resource: "user", Action: "read"},
			},
			setup: func(mockPool pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "resource", "action"}).
					AddRow(int64(100), "user", "read")
				mockPool.ExpectQuery(`SELECT id, resource, action FROM "permission" WHERE \(resource, action\) IN \(\(\$1, \$2\)\)`).
					WithArgs("user", "read").
					WillReturnRows(rows)
			},
			want: map[param.PermissionMapResourceAction]int64{
				{Resource: "user", Action: "read"}: 100,
			},
			wantErr: false,
		},
		{
			name: "Happy Path - multiple resource actions",
			resourceActions: []param.PermissionMapResourceAction{
				{Resource: "user", Action: "read"},
				{Resource: "user", Action: "write"},
				{Resource: "post", Action: "delete"},
			},
			setup: func(mockPool pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "resource", "action"}).
					AddRow(int64(100), "user", "read").
					AddRow(int64(200), "user", "write").
					AddRow(int64(300), "post", "delete")
				mockPool.ExpectQuery(`SELECT id, resource, action FROM "permission" WHERE \(resource, action\) IN`).
					WithArgs("user", "read", "user", "write", "post", "delete").
					WillReturnRows(rows)
			},
			want: map[param.PermissionMapResourceAction]int64{
				{Resource: "user", Action: "read"}:   100,
				{Resource: "user", Action: "write"}:  200,
				{Resource: "post", Action: "delete"}: 300,
			},
			wantErr: false,
		},
		{
			name: "Error - permission not found",
			resourceActions: []param.PermissionMapResourceAction{
				{Resource: "nonexistent", Action: "read"},
			},
			setup: func(mockPool pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "resource", "action"})
				mockPool.ExpectQuery(`SELECT id, resource, action FROM "permission" WHERE \(resource, action\) IN`).
					WithArgs("nonexistent", "read").
					WillReturnRows(rows)
			},
			want:    nil,
			wantErr: true,
			errMsg:  "permission not found",
		},
		{
			name:            "Happy Path - empty slice returns empty map",
			resourceActions: []param.PermissionMapResourceAction{},
			setup: func(mockPool pgxmock.PgxPoolIface) {
				// No query expected for empty input
			},
			want:    map[param.PermissionMapResourceAction]int64{},
			wantErr: false,
		},
		{
			name: "Error - partial match (one found, one not found)",
			resourceActions: []param.PermissionMapResourceAction{
				{Resource: "user", Action: "read"},
				{Resource: "user", Action: "write"},
			},
			setup: func(mockPool pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "resource", "action"}).
					AddRow(int64(100), "user", "read")
				mockPool.ExpectQuery(`SELECT id, resource, action FROM "permission" WHERE \(resource, action\) IN`).
					WithArgs("user", "read", "user", "write").
					WillReturnRows(rows)
			},
			want:    nil,
			wantErr: true,
			errMsg:  "permission not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, _ := pgxmock.NewPool()
			defer mockPool.Close()

			s := miniredis.RunT(t)
			redisClient := redis.NewClient(&redis.Options{
				Addr: s.Addr(),
			})
			defer redisClient.Close()

			tt.setup(mockPool)

			resolver := NewPermissionResolver(mockPool, redisClient, "permission", 0, infra.NewNoopLogger(), infra.NewNoopTracer())

			got, err := resolver.IDsByResourceActions(context.Background(), tt.resourceActions)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestPermissionResolverResourceActionsByIDs(t *testing.T) {
	tests := []struct {
		name    string
		ids     []int64
		setup   func(pgxmock.PgxPoolIface)
		want    map[int64]param.PermissionMapResourceAction
		wantErr bool
		errMsg  string
	}{
		{
			name: "Happy Path - single ID",
			ids:  []int64{100},
			setup: func(mockPool pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "resource", "action"}).
					AddRow(int64(100), "user", "read")
				mockPool.ExpectQuery(`SELECT id, resource, action FROM "permission" WHERE id = ANY\(\$1\)`).
					WithArgs([]int64{100}).
					WillReturnRows(rows)
			},
			want: map[int64]param.PermissionMapResourceAction{
				100: {Resource: "user", Action: "read"},
			},
			wantErr: false,
		},
		{
			name: "Happy Path - multiple IDs",
			ids:  []int64{100, 200, 300},
			setup: func(mockPool pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "resource", "action"}).
					AddRow(int64(100), "user", "read").
					AddRow(int64(200), "user", "write").
					AddRow(int64(300), "post", "delete")
				mockPool.ExpectQuery(`SELECT id, resource, action FROM "permission" WHERE id = ANY\(\$1\)`).
					WithArgs([]int64{100, 200, 300}).
					WillReturnRows(rows)
			},
			want: map[int64]param.PermissionMapResourceAction{
				100: {Resource: "user", Action: "read"},
				200: {Resource: "user", Action: "write"},
				300: {Resource: "post", Action: "delete"},
			},
			wantErr: false,
		},
		{
			name: "Error - permission not found",
			ids:  []int64{999},
			setup: func(mockPool pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "resource", "action"})
				mockPool.ExpectQuery(`SELECT id, resource, action FROM "permission" WHERE id = ANY\(\$1\)`).
					WithArgs([]int64{999}).
					WillReturnRows(rows)
			},
			want:    nil,
			wantErr: true,
			errMsg:  "permission not found",
		},
		{
			name: "Happy Path - empty slice returns empty map",
			ids:  []int64{},
			setup: func(mockPool pgxmock.PgxPoolIface) {
				// No query expected for empty input
			},
			want:    map[int64]param.PermissionMapResourceAction{},
			wantErr: false,
		},
		{
			name: "Error - partial match (one found, one not found)",
			ids:  []int64{100, 999},
			setup: func(mockPool pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "resource", "action"}).
					AddRow(int64(100), "user", "read")
				mockPool.ExpectQuery(`SELECT id, resource, action FROM "permission" WHERE id = ANY\(\$1\)`).
					WithArgs([]int64{100, 999}).
					WillReturnRows(rows)
			},
			want:    nil,
			wantErr: true,
			errMsg:  "permission not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, _ := pgxmock.NewPool()
			defer mockPool.Close()

			s := miniredis.RunT(t)
			redisClient := redis.NewClient(&redis.Options{
				Addr: s.Addr(),
			})
			defer redisClient.Close()

			tt.setup(mockPool)

			resolver := NewPermissionResolver(mockPool, redisClient, "permission", 0, infra.NewNoopLogger(), infra.NewNoopTracer())

			got, err := resolver.ResourceActionsByIDs(context.Background(), tt.ids)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
