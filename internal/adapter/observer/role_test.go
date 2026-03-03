package observer

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/adityakw90/go-monitoring"
	"github.com/adityakw90/service-access/internal/core/domain/signal"
)

func TestAdapter_Observer_NewRoleObserver(t *testing.T) {
	tests := []struct {
		name    string
		logger  *mockLogger
		tracer  monitoring.Tracer
		wantNil bool
	}{
		{
			name:    "create role observer with all parameters",
			logger:  newMockLogger(),
			tracer:  newMockTracer(),
			wantNil: false,
		},
		{
			name:    "create role observer with nil logger",
			logger:  nil,
			tracer:  newMockTracer(),
			wantNil: false,
		},
		{
			name:    "create role observer with nil tracer",
			logger:  newMockLogger(),
			tracer:  nil,
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obs := NewRoleObserver(tt.logger, tt.tracer)

			if (obs == nil) != tt.wantNil {
				t.Errorf("NewRoleObserver() = %v, wantNil %v", obs, tt.wantNil)
			}
		})
	}
}

func TestAdapter_Observer_NewRolePermissionObserver(t *testing.T) {
	tests := []struct {
		name    string
		logger  *mockLogger
		tracer  monitoring.Tracer
		wantNil bool
	}{
		{
			name:    "create role permission observer with all parameters",
			logger:  newMockLogger(),
			tracer:  newMockTracer(),
			wantNil: false,
		},
		{
			name:    "create role permission observer with nil logger",
			logger:  nil,
			tracer:  newMockTracer(),
			wantNil: false,
		},
		{
			name:    "create role permission observer with nil tracer",
			logger:  newMockLogger(),
			tracer:  nil,
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obs := NewRolePermissionObserver(tt.logger, tt.tracer)

			if (obs == nil) != tt.wantNil {
				t.Errorf("NewRolePermissionObserver() = %v, wantNil %v", obs, tt.wantNil)
			}
		})
	}
}

func TestAdapter_Observer_RoleObserver_OnSignal_Success(t *testing.T) {
	now := time.Now()
	past := now.Add(-24 * time.Hour)

	tests := []struct {
		name     string
		sig      signal.SignalType
		data     signal.SignalRole
		wantKeys []string
		wantVals map[string]any
	}{
		{
			name: "role signal with operation only",
			sig:  signal.SignalStart,
			data: signal.SignalRole{
				Operation: "create",
			},
			wantKeys: []string{"signal", "operation"},
			wantVals: map[string]any{
				"operation": "create",
			},
		},
		{
			name: "role signal with uid and group_uid",
			sig:  signal.SignalSuccess,
			data: signal.SignalRole{
				Operation: "update",
				UID:        stringPtr("role-123"),
				GroupUID:   stringPtr("group-456"),
			},
			wantKeys: []string{"signal", "operation", "role.uid", "role.group_uid"},
			wantVals: map[string]any{
				"operation":     "update",
				"role.uid":     "role-123",
				"role.group_uid": "group-456",
			},
		},
		{
			name: "role signal with all fields",
			sig:  signal.SignalStart,
			data: signal.SignalRole{
				Operation:  "delete",
				UID:        stringPtr("role-789"),
				GroupUID:   stringPtr("group-abc"),
				Name:       stringPtr("editors"),
				Description: stringPtr("content editors role"),
				CreatedAt:  &past,
				UpdatedAt:  &now,
			},
			wantKeys: []string{"signal", "operation", "role.uid", "role.group_uid", "role.name", "role.description", "role.created_at", "role.updated_at"},
			wantVals: map[string]any{
				"operation":         "delete",
				"role.uid":        "role-789",
				"role.group_uid":  "group-abc",
				"role.name":       "editors",
				"role.description": "content editors role",
			},
		},
		{
			name: "role signal with timestamps",
			sig:  signal.SignalReject,
			data: signal.SignalRole{
				Operation: "validate",
				CreatedAt:  &now,
				UpdatedAt:  &past,
			},
			wantKeys: []string{"signal", "operation", "role.created_at", "role.updated_at"},
			wantVals: map[string]any{
				"operation": "validate",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := newMockLogger()
			tracer := newMockTracer()

			obs := NewRoleObserver(logger, tracer)
			ctx := context.Background()

			obs.OnSignal(ctx, tt.sig, tt.data, nil)

			// Verify debug log was called
			if len(logger.debugMessages) != 1 {
				t.Fatalf("expected 1 debug log entry, got %d", len(logger.debugMessages))
			}

			entry := logger.debugMessages[0]
			if entry.message != "service signal" {
				t.Errorf("message = %s, want 'service signal'", entry.message)
			}

			// Verify all expected keys are present
			for _, key := range tt.wantKeys {
				if _, ok := entry.fields[key]; !ok {
					t.Errorf("expected key %q not found in fields: %v", key, entry.fields)
				}
			}

			// Verify expected values
			for key, wantVal := range tt.wantVals {
				gotVal := entry.fields[key]
				if gotVal != wantVal {
					t.Errorf("key %q = %v, want %v", key, gotVal, wantVal)
				}
			}

			// Verify signal value
			if entry.fields["signal"] != tt.sig {
				t.Errorf("signal = %v, want %v", entry.fields["signal"], tt.sig)
			}
		})
	}
}

func TestAdapter_Observer_RolePermissionObserver_OnSignal_Success(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		sig      signal.SignalType
		data     signal.SignalRolePermission
		wantKeys []string
		wantVals map[string]any
	}{
		{
			name: "role permission signal with operation only",
			sig:  signal.SignalStart,
			data: signal.SignalRolePermission{
				Operation: "attach",
			},
			wantKeys: []string{"signal", "operation"},
			wantVals: map[string]any{
				"operation": "attach",
			},
		},
		{
			name: "role permission signal with role_uid and permission_uid",
			sig:  signal.SignalSuccess,
			data: signal.SignalRolePermission{
				Operation:     "detach",
				RoleUID:       stringPtr("role-123"),
				PermissionUID: stringPtr("perm-456"),
			},
			wantKeys: []string{"signal", "operation", "role_permission.group_uid", "role_permission.permission_uid"},
			wantVals: map[string]any{
				"operation":                          "detach",
				"role_permission.group_uid":        "role-123",
				"role_permission.permission_uid":   "perm-456",
			},
		},
		{
			name: "role permission signal with permission details",
			sig:  signal.SignalStart,
			data: signal.SignalRolePermission{
				Operation:             "attach",
				RoleUID:               stringPtr("role-789"),
				GroupPermissionUID:    stringPtr("gp-abc"),
				PermissionUID:         stringPtr("perm-def"),
				PermissionResource:    stringPtr("documents"),
				PermissionAction:      stringPtr("write"),
				PermissionDescription: stringPtr("document write permission"),
				CreatedAt:             &now,
			},
			wantKeys: []string{"signal", "operation", "role_permission.group_uid", "role_permission.group_permission_uid", "role_permission.permission_uid", "role_permission.permission_resource", "role_permission.permission_action", "role_permission.permission_description", "role_permission.created_at"},
			wantVals: map[string]any{
				"operation":                                "attach",
				"role_permission.group_uid":              "role-789",
				"role_permission.group_permission_uid":   "gp-abc",
				"role_permission.permission_uid":         "perm-def",
				"role_permission.permission_resource":    "documents",
				"role_permission.permission_action":      "write",
				"role_permission.permission_description": "document write permission",
			},
		},
		{
			name: "role permission signal with group_permission_uids array",
			sig:  signal.SignalReject,
			data: signal.SignalRolePermission{
				Operation:           "bulk-attach",
				RoleUID:             stringPtr("role-123"),
				GroupPermissionUIDs: []string{"gp-1", "gp-2", "gp-3"},
			},
			wantKeys: []string{"signal", "operation", "role_permission.group_uid", "role_permission.group_permission_uids"},
			wantVals: map[string]any{
				"operation":                  "bulk-attach",
				"role_permission.group_uid": "role-123",
			},
		},
		{
			name: "role permission signal with permission_uids array",
			sig:  signal.SignalFail,
			data: signal.SignalRolePermission{
				Operation:      "bulk-detach",
				RoleUID:        stringPtr("role-456"),
				PermissionUIDs: []string{"perm-1", "perm-2", "perm-3"},
			},
			wantKeys: []string{"signal", "operation", "role_permission.group_uid", "role_permission.permission_uids"},
			wantVals: map[string]any{
				"operation":                  "bulk-detach",
				"role_permission.group_uid": "role-456",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := newMockLogger()
			tracer := newMockTracer()

			obs := NewRolePermissionObserver(logger, tracer)
			ctx := context.Background()

			obs.OnSignal(ctx, tt.sig, tt.data, nil)

			// Verify debug log was called
			if len(logger.debugMessages) != 1 {
				t.Fatalf("expected 1 debug log entry, got %d", len(logger.debugMessages))
			}

			entry := logger.debugMessages[0]
			if entry.message != "service signal" {
				t.Errorf("message = %s, want 'service signal'", entry.message)
			}

			// Verify all expected keys are present
			for _, key := range tt.wantKeys {
				if _, ok := entry.fields[key]; !ok {
					t.Errorf("expected key %q not found in fields: %v", key, entry.fields)
				}
			}

			// Verify expected values
			for key, wantVal := range tt.wantVals {
				gotVal := entry.fields[key]
				if gotVal != wantVal {
					t.Errorf("key %q = %v, want %v", key, gotVal, wantVal)
				}
			}

			// Verify signal value
			if entry.fields["signal"] != tt.sig {
				t.Errorf("signal = %v, want %v", entry.fields["signal"], tt.sig)
			}
		})
	}
}

func TestAdapter_Observer_RoleObserver_OnSignal_WithError(t *testing.T) {
	tests := []struct {
		name    string
		sig     signal.SignalType
		data    signal.SignalRole
		err     error
		wantMsg string
	}{
		{
			name: "role with database error",
			sig:  signal.SignalFail,
			data: signal.SignalRole{
				Operation: "create",
				UID:        stringPtr("role-123"),
			},
			err:     errors.New("database connection failed"),
			wantMsg: "service signal",
		},
		{
			name: "role with validation error",
			sig:  signal.SignalReject,
			data: signal.SignalRole{
				Operation: "update",
				Name:       stringPtr(""),
			},
			err:     errors.New("invalid role name"),
			wantMsg: "service signal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := newMockLogger()
			tracer := newMockTracer()

			obs := NewRoleObserver(logger, tracer)
			ctx := context.Background()

			obs.OnSignal(ctx, tt.sig, tt.data, tt.err)

			// Verify error log was called
			if len(logger.errorMessages) != 1 {
				t.Fatalf("expected 1 error log entry, got %d", len(logger.errorMessages))
			}

			entry := logger.errorMessages[0]
			if entry.message != tt.wantMsg {
				t.Errorf("message = %s, want %s", entry.message, tt.wantMsg)
			}

			// Verify error field is present
			if _, ok := entry.fields["error.Type"]; !ok {
				t.Errorf("error.Type field not found")
			}

			if _, ok := entry.fields["error.Message"]; !ok {
				t.Errorf("error.Message field not found")
			}

			if entry.fields["error.Message"] != tt.err.Error() {
				t.Errorf("error.Message = %v, want %v", entry.fields["error.Message"], tt.err.Error())
			}

			// Verify signal value
			if entry.fields["signal"] != tt.sig {
				t.Errorf("signal = %v, want %v", entry.fields["signal"], tt.sig)
			}

			// Verify operation is preserved in error logs
			if entry.fields["operation"] != tt.data.Operation {
				t.Errorf("operation = %v, want %v", entry.fields["operation"], tt.data.Operation)
			}
		})
	}
}

func TestAdapter_Observer_RolePermissionObserver_OnSignal_WithError(t *testing.T) {
	tests := []struct {
		name    string
		sig     signal.SignalType
		data    signal.SignalRolePermission
		err     error
		wantMsg string
	}{
		{
			name: "role permission with database error",
			sig:  signal.SignalFail,
			data: signal.SignalRolePermission{
				Operation:     "attach",
				RoleUID:       stringPtr("role-123"),
				PermissionUID: stringPtr("perm-456"),
			},
			err:     errors.New("permission not found"),
			wantMsg: "service signal",
		},
		{
			name: "role permission with validation error",
			sig:  signal.SignalReject,
			data: signal.SignalRolePermission{
				Operation: "detach",
				RoleUID:   stringPtr(""),
			},
			err:     errors.New("invalid role UID"),
			wantMsg: "service signal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := newMockLogger()
			tracer := newMockTracer()

			obs := NewRolePermissionObserver(logger, tracer)
			ctx := context.Background()

			obs.OnSignal(ctx, tt.sig, tt.data, tt.err)

			// Verify error log was called
			if len(logger.errorMessages) != 1 {
				t.Fatalf("expected 1 error log entry, got %d", len(logger.errorMessages))
			}

			entry := logger.errorMessages[0]
			if entry.message != tt.wantMsg {
				t.Errorf("message = %s, want %s", entry.message, tt.wantMsg)
			}

			// Verify error field is present
			if _, ok := entry.fields["error.Type"]; !ok {
				t.Errorf("error.Type field not found")
			}

			if _, ok := entry.fields["error.Message"]; !ok {
				t.Errorf("error.Message field not found")
			}

			if entry.fields["error.Message"] != tt.err.Error() {
				t.Errorf("error.Message = %v, want %v", entry.fields["error.Message"], tt.err.Error())
			}

			// Verify signal value
			if entry.fields["signal"] != tt.sig {
				t.Errorf("signal = %v, want %v", entry.fields["signal"], tt.sig)
			}
		})
	}
}

func TestAdapter_Observer_RoleObserver_AllSignalTypes(t *testing.T) {
	signals := []signal.SignalType{
		signal.SignalStart,
		signal.SignalReject,
		signal.SignalFail,
		signal.SignalSuccess,
	}

	for _, sig := range signals {
		t.Run(string(sig), func(t *testing.T) {
			logger := newMockLogger()
			tracer := newMockTracer()

			obs := NewRoleObserver(logger, tracer)
			ctx := context.Background()

			data := signal.SignalRole{
				Operation:  "test-operation",
				UID:        stringPtr("test-role"),
				GroupUID:   stringPtr("test-group"),
				Name:       stringPtr("test-name"),
				Description: stringPtr("test description"),
			}
			obs.OnSignal(ctx, sig, data, nil)

			// Verify debug log was called
			if len(logger.debugMessages) != 1 {
				t.Fatalf("expected 1 debug log entry, got %d", len(logger.debugMessages))
			}

			entry := logger.debugMessages[0]
			if entry.fields["signal"] != sig {
				t.Errorf("signal = %v, want %v", entry.fields["signal"], sig)
			}
		})
	}
}

func TestAdapter_Observer_RolePermissionObserver_AllSignalTypes(t *testing.T) {
	signals := []signal.SignalType{
		signal.SignalStart,
		signal.SignalReject,
		signal.SignalFail,
		signal.SignalSuccess,
	}

	for _, sig := range signals {
		t.Run(string(sig), func(t *testing.T) {
			logger := newMockLogger()
			tracer := newMockTracer()

			obs := NewRolePermissionObserver(logger, tracer)
			ctx := context.Background()

			data := signal.SignalRolePermission{
				Operation:          "test-operation",
				RoleUID:            stringPtr("test-role"),
				GroupPermissionUID: stringPtr("test-gp"),
				PermissionUID:      stringPtr("test-perm"),
				PermissionResource: stringPtr("test-resource"),
				PermissionAction:   stringPtr("test-action"),
			}
			obs.OnSignal(ctx, sig, data, nil)

			// Verify debug log was called
			if len(logger.debugMessages) != 1 {
				t.Fatalf("expected 1 debug log entry, got %d", len(logger.debugMessages))
			}

			entry := logger.debugMessages[0]
			if entry.fields["signal"] != sig {
				t.Errorf("signal = %v, want %v", entry.fields["signal"], sig)
			}
		})
	}
}

func TestAdapter_Observer_RoleObserver_NilFields(t *testing.T) {
	logger := newMockLogger()
	tracer := newMockTracer()

	obs := NewRoleObserver(logger, tracer)
	ctx := context.Background()

	// All nil fields except operation
	data := signal.SignalRole{
		Operation: "nil-test",
	}
	obs.OnSignal(ctx, signal.SignalStart, data, nil)

	if len(logger.debugMessages) != 1 {
		t.Fatalf("expected 1 debug log entry, got %d", len(logger.debugMessages))
	}

	entry := logger.debugMessages[0]
	// Should only have signal and operation
	if entry.fields["operation"] != "nil-test" {
		t.Errorf("operation = %v, want 'nil-test'", entry.fields["operation"])
	}

	// These keys should not be present since they were nil
	nilKeys := []string{"role.uid", "role.group_uid", "role.name", "role.description", "role.created_at", "role.updated_at"}
	for _, key := range nilKeys {
		if _, ok := entry.fields[key]; ok {
			t.Errorf("key %q should not be present when nil", key)
		}
	}
}

func TestAdapter_Observer_RolePermissionObserver_NilFields(t *testing.T) {
	logger := newMockLogger()
	tracer := newMockTracer()

	obs := NewRolePermissionObserver(logger, tracer)
	ctx := context.Background()

	// All nil fields except operation
	data := signal.SignalRolePermission{
		Operation: "nil-test",
	}
	obs.OnSignal(ctx, signal.SignalStart, data, nil)

	if len(logger.debugMessages) != 1 {
		t.Fatalf("expected 1 debug log entry, got %d", len(logger.debugMessages))
	}

	entry := logger.debugMessages[0]
	// Should only have signal and operation
	if entry.fields["operation"] != "nil-test" {
		t.Errorf("operation = %v, want 'nil-test'", entry.fields["operation"])
	}

	// These keys should not be present since they were nil
	nilKeys := []string{"role_permission.group_uid", "role_permission.group_permission_uid", "role_permission.permission_uid", "role_permission.permission_resource", "role_permission.permission_action", "role_permission.permission_description", "role_permission.created_at", "role_permission.group_permission_uids", "role_permission.permission_uids"}
	for _, key := range nilKeys {
		if _, ok := entry.fields[key]; ok {
			t.Errorf("key %q should not be present when nil", key)
		}
	}
}

func TestAdapter_Observer_RolePermissionObserver_GroupPermissionUIDs(t *testing.T) {
	logger := newMockLogger()
	tracer := newMockTracer()

	obs := NewRolePermissionObserver(logger, tracer)
	ctx := context.Background()

	data := signal.SignalRolePermission{
		Operation:           "bulk-attach",
		RoleUID:             stringPtr("role-123"),
		GroupPermissionUIDs: []string{"gp-1", "gp-2", "gp-3"},
	}
	obs.OnSignal(ctx, signal.SignalStart, data, nil)

	if len(logger.debugMessages) != 1 {
		t.Fatalf("expected 1 debug log entry, got %d", len(logger.debugMessages))
	}

	entry := logger.debugMessages[0]
	// Verify group_permission_uids is present
	if _, ok := entry.fields["role_permission.group_permission_uids"]; !ok {
		t.Errorf("role_permission.group_permission_uids field not found")
	}
}

func TestAdapter_Observer_RolePermissionObserver_PermissionUIDs(t *testing.T) {
	logger := newMockLogger()
	tracer := newMockTracer()

	obs := NewRolePermissionObserver(logger, tracer)
	ctx := context.Background()

	data := signal.SignalRolePermission{
		Operation:      "bulk-attach",
		RoleUID:        stringPtr("role-123"),
		PermissionUIDs: []string{"perm-1", "perm-2", "perm-3"},
	}
	obs.OnSignal(ctx, signal.SignalStart, data, nil)

	if len(logger.debugMessages) != 1 {
		t.Fatalf("expected 1 debug log entry, got %d", len(logger.debugMessages))
	}

	entry := logger.debugMessages[0]
	// Verify permission_uids is present
	if _, ok := entry.fields["role_permission.permission_uids"]; !ok {
		t.Errorf("role_permission.permission_uids field not found")
	}
}
