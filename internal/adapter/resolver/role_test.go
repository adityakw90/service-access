package resolver

import (
	"context"
	"testing"

	"github.com/adityakw90/service-access/internal/infra"
	"github.com/alicebob/miniredis/v2"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoleResolverIDsByUIDs(t *testing.T) {
	tests := []struct {
		name    string
		uids    []string
		setup   func(pgxmock.PgxPoolIface)
		want    map[string]int64
		wantErr bool
		errMsg  string
	}{
		{
			name: "Happy Path - single UID",
			uids: []string{"role-uid-1"},
			setup: func(mockPool pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "uid"}).
					AddRow(int64(100), "role-uid-1")
				mockPool.ExpectQuery(`SELECT id, uid FROM "role" WHERE uid=\$1`).
					WithArgs("role-uid-1").
					WillReturnRows(rows)
			},
			want:    map[string]int64{"role-uid-1": 100},
			wantErr: false,
		},
		{
			name: "Happy Path - multiple UIDs",
			uids: []string{"role-uid-1", "role-uid-2"},
			setup: func(mockPool pgxmock.PgxPoolIface) {
				// Each UID is fetched separately in goroutines
				// We use ExpectQueryArranged to allow any order of execution
				rows1 := pgxmock.NewRows([]string{"id", "uid"}).
					AddRow(int64(100), "role-uid-1")
				mockPool.ExpectQuery(`SELECT id, uid FROM "role" WHERE uid=\$1`).
					WithArgs("role-uid-1").
					WillReturnRows(rows1)

				rows2 := pgxmock.NewRows([]string{"id", "uid"}).
					AddRow(int64(200), "role-uid-2")
				mockPool.ExpectQuery(`SELECT id, uid FROM "role" WHERE uid=\$1`).
					WithArgs("role-uid-2").
					WillReturnRows(rows2)

				// Allow expectations to be satisfied in any order
				mockPool.MatchExpectationsInOrder(false)
			},
			want:    map[string]int64{"role-uid-1": 100, "role-uid-2": 200},
			wantErr: false,
		},
		{
			name: "Error - role not found",
			uids: []string{"non-existent"},
			setup: func(mockPool pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "uid"})
				mockPool.ExpectQuery(`SELECT id, uid FROM "role" WHERE uid=\$1`).
					WithArgs("non-existent").
					WillReturnRows(rows)
			},
			want:    nil,
			wantErr: true,
			errMsg:  "role not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, _ := pgxmock.NewPool()
			defer mockPool.Close()

			// Setup miniredis
			s := miniredis.RunT(t)
			redisClient := redis.NewClient(&redis.Options{
				Addr: s.Addr(),
			})
			defer redisClient.Close()

			tt.setup(mockPool)

			resolver := NewRoleResolver(mockPool, redisClient, "role", 0, infra.NewNoopLogger(), infra.NewNoopTracer())

			got, err := resolver.IDsByUIDs(context.Background(), tt.uids)

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

func TestRoleResolverUIDsByIDs(t *testing.T) {
	tests := []struct {
		name    string
		ids     []int64
		setup   func(pgxmock.PgxPoolIface)
		want    map[int64]string
		wantErr bool
		errMsg  string
	}{
		{
			name: "Happy Path - single ID",
			ids:  []int64{100},
			setup: func(mockPool pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "uid"}).
					AddRow(int64(100), "role-uid-1")
				mockPool.ExpectQuery(`SELECT id, uid FROM "role" WHERE id=\$1`).
					WithArgs(int64(100)).
					WillReturnRows(rows)
			},
			want:    map[int64]string{100: "role-uid-1"},
			wantErr: false,
		},
		{
			name: "Happy Path - multiple IDs",
			ids:  []int64{100, 200},
			setup: func(mockPool pgxmock.PgxPoolIface) {
				// Each ID is fetched separately in goroutines
				rows1 := pgxmock.NewRows([]string{"id", "uid"}).
					AddRow(int64(100), "role-uid-1")
				mockPool.ExpectQuery(`SELECT id, uid FROM "role" WHERE id=\$1`).
					WithArgs(int64(100)).
					WillReturnRows(rows1)

				rows2 := pgxmock.NewRows([]string{"id", "uid"}).
					AddRow(int64(200), "role-uid-2")
				mockPool.ExpectQuery(`SELECT id, uid FROM "role" WHERE id=\$1`).
					WithArgs(int64(200)).
					WillReturnRows(rows2)

				// Allow expectations to be satisfied in any order
				mockPool.MatchExpectationsInOrder(false)
			},
			want:    map[int64]string{100: "role-uid-1", 200: "role-uid-2"},
			wantErr: false,
		},
		{
			name: "Error - role not found",
			ids:  []int64{999},
			setup: func(mockPool pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "uid"})
				mockPool.ExpectQuery(`SELECT id, uid FROM "role" WHERE id=\$1`).
					WithArgs(int64(999)).
					WillReturnRows(rows)
			},
			want:    nil,
			wantErr: true,
			errMsg:  "role not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, _ := pgxmock.NewPool()
			defer mockPool.Close()

			// Setup miniredis
			s := miniredis.RunT(t)
			redisClient := redis.NewClient(&redis.Options{
				Addr: s.Addr(),
			})
			defer redisClient.Close()

			tt.setup(mockPool)

			resolver := NewRoleResolver(mockPool, redisClient, "role", 0, infra.NewNoopLogger(), infra.NewNoopTracer())

			got, err := resolver.UIDsByIDs(context.Background(), tt.ids)

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
