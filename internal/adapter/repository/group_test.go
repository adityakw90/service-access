package repository

import (
	"context"
	"testing"
	"time"

	"github.com/adityakw90/service-access/internal/core/domain/model"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdapter_GroupRepository_Create(t *testing.T) {
	tests := []struct {
		name    string
		data    map[string]any
		wantErr bool
	}{
		{
			name: "Create valid group",
			data: map[string]any{
				"uid":         "group-uid-001",
				"name":        "group-name-001",
				"description": "group-description-001",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			repo := NewGroupRepository(mockPool)
			group := &model.Group{
				UID:         tt.data["uid"].(string),
				Name:        tt.data["name"].(string),
				Description: tt.data["description"].(string),
			}

			rows := pgxmock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow(int64(1), time.Time{}, time.Time{})

			mockPool.ExpectQuery(`
				INSERT INTO "group" \(uid, name, description\)
				VALUES \(\$1, \$2, \$3\)
				RETURNING id, created_at, updated_at
			`).
				WithArgs(group.UID, group.Name, group.Description).
				WillReturnRows(rows)

			err = repo.Create(context.Background(), group)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.data["uid"], group.UID)
			assert.Equal(t, tt.data["name"], group.Name)
			assert.Equal(t, tt.data["description"], group.Description)
			assert.Equal(t, int64(1), group.ID)
			assert.Equal(t, time.Time{}, group.CreatedAt)
			assert.Equal(t, time.Time{}, group.UpdatedAt)
			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}
