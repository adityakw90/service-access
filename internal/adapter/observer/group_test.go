package observer

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/adityakw90/go-monitoring"
	"github.com/adityakw90/service-access/internal/core/domain/signal"
)

func TestAdapter_Observer_NewGroupObserver(t *testing.T) {
	tests := []struct {
		name    string
		logger  *mockLogger
		tracer  monitoring.Tracer
		wantNil bool
	}{
		{
			name:    "create group observer with all parameters",
			logger:  newMockLogger(),
			tracer:  newMockTracer(),
			wantNil: false,
		},
		{
			name:    "create group observer with nil logger",
			logger:  nil,
			tracer:  newMockTracer(),
			wantNil: false,
		},
		{
			name:    "create group observer with nil tracer",
			logger:  newMockLogger(),
			tracer:  nil,
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obs := NewGroupObserver(tt.logger, tt.tracer)

			if (obs == nil) != tt.wantNil {
				t.Errorf("NewGroupObserver() = %v, wantNil %v", obs, tt.wantNil)
			}
		})
	}
}

func TestAdapter_Observer_NewGroupPermissionObserver(t *testing.T) {
	tests := []struct {
		name    string
		logger  *mockLogger
		tracer  monitoring.Tracer
		wantNil bool
	}{
		{
			name:    "create group permission observer with all parameters",
			logger:  newMockLogger(),
			tracer:  newMockTracer(),
			wantNil: false,
		},
		{
			name:    "create group permission observer with nil logger",
			logger:  nil,
			tracer:  newMockTracer(),
			wantNil: false,
		},
		{
			name:    "create group permission observer with nil tracer",
			logger:  newMockLogger(),
			tracer:  nil,
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obs := NewGroupPermissionObserver(tt.logger, tt.tracer)

			if (obs == nil) != tt.wantNil {
				t.Errorf("NewGroupPermissionObserver() = %v, wantNil %v", obs, tt.wantNil)
			}
		})
	}
}

func TestAdapter_Observer_GroupObserver_OnSignal_Success(t *testing.T) {
	now := time.Now()
	past := now.Add(-24 * time.Hour)

	tests := []struct {
		name     string
		sig      signal.SignalType
		data     signal.SignalGroup
		wantKeys []string
		wantVals map[string]any
	}{
		{
			name: "group signal with operation only",
			sig:  signal.SignalStart,
			data: signal.SignalGroup{
				Operation: "create",
			},
			wantKeys: []string{"signal", "operation"},
			wantVals: map[string]any{
				"operation": "create",
			},
		},
		{
			name: "group signal with uid and name",
			sig:  signal.SignalSuccess,
			data: signal.SignalGroup{
				Operation: "update",
				UID:        stringPtr("group-123"),
				Name:       stringPtr("admins"),
			},
			wantKeys: []string{"signal", "operation", "group.uid", "group.name"},
			wantVals: map[string]any{
				"operation":  "update",
				"group.uid":  "group-123",
				"group.name": "admins",
			},
		},
		{
			name: "group signal with all fields",
			sig:  signal.SignalStart,
			data: signal.SignalGroup{
				Operation:  "delete",
				UID:        stringPtr("group-456"),
				Name:       stringPtr("editors"),
				Description: stringPtr("content editors group"),
				CreatedAt:  &past,
				UpdatedAt:  &now,
			},
			wantKeys: []string{"signal", "operation", "group.uid", "group.name", "group.description", "group.created_at", "group.updated_at"},
			wantVals: map[string]any{
				"operation":        "delete",
				"group.uid":       "group-456",
				"group.name":      "editors",
				"group.description": "content editors group",
			},
		},
		{
			name: "group signal with timestamps",
			sig:  signal.SignalReject,
			data: signal.SignalGroup{
				Operation: "validate",
				CreatedAt:  &now,
				UpdatedAt:  &past,
			},
			wantKeys: []string{"signal", "operation", "group.created_at", "group.updated_at"},
			wantVals: map[string]any{
				"operation": "validate",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := newMockLogger()
			tracer := newMockTracer()

			obs := NewGroupObserver(logger, tracer)
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

func TestAdapter_Observer_GroupPermissionObserver_OnSignal_Success(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		sig      signal.SignalType
		data     signal.SignalGroupPermission
		wantKeys []string
		wantVals map[string]any
	}{
		{
			name: "group permission signal with operation only",
			sig:  signal.SignalStart,
			data: signal.SignalGroupPermission{
				Operation: "attach",
			},
			wantKeys: []string{"signal", "operation"},
			wantVals: map[string]any{
				"operation": "attach",
			},
		},
		{
			name: "group permission signal with group_uid and permission_uid",
			sig:  signal.SignalSuccess,
			data: signal.SignalGroupPermission{
				Operation:      "detach",
				UID:            stringPtr("gp-123"),
				GroupUID:       stringPtr("group-123"),
				PermissionUID:  stringPtr("perm-456"),
			},
			wantKeys: []string{"signal", "operation", "group_permission.uid", "group_permission.group_uid", "group_permission.permission_uid"},
			wantVals: map[string]any{
				"operation":                       "detach",
				"group_permission.uid":            "gp-123",
				"group_permission.group_uid":      "group-123",
				"group_permission.permission_uid": "perm-456",
			},
		},
		{
			name: "group permission signal with permission details",
			sig:  signal.SignalStart,
			data: signal.SignalGroupPermission{
				Operation:             "attach",
				UID:                   stringPtr("gp-789"),
				GroupUID:              stringPtr("group-abc"),
				PermissionUID:         stringPtr("perm-def"),
				PermissionResource:    stringPtr("documents"),
				PermissionAction:      stringPtr("write"),
				PermissionDescription: stringPtr("document write permission"),
				CreatedAt:             &now,
			},
			wantKeys: []string{"signal", "operation", "group_permission.uid", "group_permission.group_uid", "group_permission.permission_uid", "group_permission.permission_resource", "group_permission.permission_action", "group_permission.permission_description", "group_permission.created_at"},
			wantVals: map[string]any{
				"operation":                               "attach",
				"group_permission.uid":                   "gp-789",
				"group_permission.group_uid":             "group-abc",
				"group_permission.permission_uid":        "perm-def",
				"group_permission.permission_resource":   "documents",
				"group_permission.permission_action":     "write",
				"group_permission.permission_description": "document write permission",
			},
		},
		{
			name: "group permission signal with permission_uids array",
			sig:  signal.SignalReject,
			data: signal.SignalGroupPermission{
				Operation:      "bulk-attach",
				GroupUID:       stringPtr("group-123"),
				PermissionUIDs: []string{"perm-1", "perm-2", "perm-3"},
			},
			wantKeys: []string{"signal", "operation", "group_permission.group_uid", "group_permission.permission_uids"},
			wantVals: map[string]any{
				"operation":                  "bulk-attach",
				"group_permission.group_uid": "group-123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := newMockLogger()
			tracer := newMockTracer()

			obs := NewGroupPermissionObserver(logger, tracer)
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

func TestAdapter_Observer_GroupObserver_OnSignal_WithError(t *testing.T) {
	tests := []struct {
		name    string
		sig     signal.SignalType
		data    signal.SignalGroup
		err     error
		wantMsg string
	}{
		{
			name: "group with database error",
			sig:  signal.SignalFail,
			data: signal.SignalGroup{
				Operation: "create",
				UID:        stringPtr("group-123"),
			},
			err:     errors.New("database connection failed"),
			wantMsg: "service signal",
		},
		{
			name: "group with validation error",
			sig:  signal.SignalReject,
			data: signal.SignalGroup{
				Operation: "update",
				Name:       stringPtr(""),
			},
			err:     errors.New("invalid group name"),
			wantMsg: "service signal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := newMockLogger()
			tracer := newMockTracer()

			obs := NewGroupObserver(logger, tracer)
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

func TestAdapter_Observer_GroupPermissionObserver_OnSignal_WithError(t *testing.T) {
	tests := []struct {
		name    string
		sig     signal.SignalType
		data    signal.SignalGroupPermission
		err     error
		wantMsg string
	}{
		{
			name: "group permission with database error",
			sig:  signal.SignalFail,
			data: signal.SignalGroupPermission{
				Operation:     "attach",
				GroupUID:      stringPtr("group-123"),
				PermissionUID: stringPtr("perm-456"),
			},
			err:     errors.New("permission not found"),
			wantMsg: "service signal",
		},
		{
			name: "group permission with validation error",
			sig:  signal.SignalReject,
			data: signal.SignalGroupPermission{
				Operation: "detach",
				GroupUID:  stringPtr(""),
			},
			err:     errors.New("invalid group UID"),
			wantMsg: "service signal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := newMockLogger()
			tracer := newMockTracer()

			obs := NewGroupPermissionObserver(logger, tracer)
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

func TestAdapter_Observer_GroupObserver_AllSignalTypes(t *testing.T) {
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

			obs := NewGroupObserver(logger, tracer)
			ctx := context.Background()

			data := signal.SignalGroup{
				Operation:  "test-operation",
				UID:        stringPtr("test-group"),
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

func TestAdapter_Observer_GroupPermissionObserver_AllSignalTypes(t *testing.T) {
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

			obs := NewGroupPermissionObserver(logger, tracer)
			ctx := context.Background()

			data := signal.SignalGroupPermission{
				Operation:          "test-operation",
				GroupUID:           stringPtr("test-group"),
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

func TestAdapter_Observer_GroupObserver_NilFields(t *testing.T) {
	logger := newMockLogger()
	tracer := newMockTracer()

	obs := NewGroupObserver(logger, tracer)
	ctx := context.Background()

	// All nil fields except operation
	data := signal.SignalGroup{
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
	nilKeys := []string{"group.uid", "group.name", "group.description", "group.created_at", "group.updated_at"}
	for _, key := range nilKeys {
		if _, ok := entry.fields[key]; ok {
			t.Errorf("key %q should not be present when nil", key)
		}
	}
}

func TestAdapter_Observer_GroupPermissionObserver_NilFields(t *testing.T) {
	logger := newMockLogger()
	tracer := newMockTracer()

	obs := NewGroupPermissionObserver(logger, tracer)
	ctx := context.Background()

	// All nil fields except operation
	data := signal.SignalGroupPermission{
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
	nilKeys := []string{"group_permission.uid", "group_permission.group_uid", "group_permission.permission_uid", "group_permission.permission_resource", "group_permission.permission_action", "group_permission.permission_description", "group_permission.created_at", "group_permission.permission_uids"}
	for _, key := range nilKeys {
		if _, ok := entry.fields[key]; ok {
			t.Errorf("key %q should not be present when nil", key)
		}
	}
}

func TestAdapter_Observer_GroupPermissionObserver_PermissionUIDs(t *testing.T) {
	logger := newMockLogger()
	tracer := newMockTracer()

	obs := NewGroupPermissionObserver(logger, tracer)
	ctx := context.Background()

	data := signal.SignalGroupPermission{
		Operation:      "bulk-attach",
		GroupUID:       stringPtr("group-123"),
		PermissionUIDs: []string{"perm-1", "perm-2", "perm-3"},
	}
	obs.OnSignal(ctx, signal.SignalStart, data, nil)

	if len(logger.debugMessages) != 1 {
		t.Fatalf("expected 1 debug log entry, got %d", len(logger.debugMessages))
	}

	entry := logger.debugMessages[0]
	// Verify permission_uids is present
	if _, ok := entry.fields["group_permission.permission_uids"]; !ok {
		t.Errorf("group_permission.permission_uids field not found")
	}
}
