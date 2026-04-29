package observer

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/adityakw90/go-monitoring"
	"github.com/adityakw90/service-access/internal/core/domain/signal"
	"github.com/adityakw90/service-access/pkg/util"
)

func TestAdapter_Observer_NewPermissionObserver(t *testing.T) {
	tests := []struct {
		name    string
		logger  *mockLogger
		tracer  monitoring.Tracer
		wantNil bool
	}{
		{
			name:    "create permission observer with all parameters",
			logger:  newMockLogger(),
			tracer:  newMockTracer(),
			wantNil: false,
		},
		{
			name:    "create permission observer with nil logger",
			logger:  nil,
			tracer:  newMockTracer(),
			wantNil: false,
		},
		{
			name:    "create permission observer with nil tracer",
			logger:  newMockLogger(),
			tracer:  nil,
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obs := NewPermissionObserver(tt.logger, tt.tracer)

			if (obs == nil) != tt.wantNil {
				t.Errorf("NewPermissionObserver() = %v, wantNil %v", obs, tt.wantNil)
			}
		})
	}
}

func TestAdapter_Observer_PermissionObserver_OnSignal_Success(t *testing.T) {
	now := time.Now()
	past := now.Add(-24 * time.Hour)

	tests := []struct {
		name     string
		sig      signal.SignalType
		data     signal.SignalPermission
		wantKeys []string
		wantVals map[string]any
	}{
		{
			name: "permission signal with operation only",
			sig:  signal.SignalStart,
			data: signal.SignalPermission{
				Operation: "create",
			},
			wantKeys: []string{"signal", "operation"},
			wantVals: map[string]any{
				"operation": "create",
			},
		},
		{
			name: "permission signal with uid and resource",
			sig:  signal.SignalSuccess,
			data: signal.SignalPermission{
				Operation: "update",
				UID:       util.Ptr("perm-123"),
				Resource:  util.Ptr("documents"),
			},
			wantKeys: []string{"signal", "operation", "permission.uid", "permission.resource"},
			wantVals: map[string]any{
				"operation":           "update",
				"permission.uid":      "perm-123",
				"permission.resource": "documents",
			},
		},
		{
			name: "permission signal with all fields",
			sig:  signal.SignalStart,
			data: signal.SignalPermission{
				Operation:   "delete",
				UID:         util.Ptr("perm-456"),
				Resource:    util.Ptr("admin"),
				Action:      util.Ptr("delete"),
				Description: util.Ptr("admin delete permission"),
				CreatedAt:   &past,
				UpdatedAt:   &now,
			},
			wantKeys: []string{"signal", "operation", "permission.uid", "permission.resource", "permission.action", "permission.description", "permission.created_at", "permission.updated_at"},
			wantVals: map[string]any{
				"operation":              "delete",
				"permission.uid":         "perm-456",
				"permission.resource":    "admin",
				"permission.action":      "delete",
				"permission.description": "admin delete permission",
			},
		},
		{
			name: "permission signal with timestamps",
			sig:  signal.SignalReject,
			data: signal.SignalPermission{
				Operation: "validate",
				CreatedAt: &now,
				UpdatedAt: &past,
			},
			wantKeys: []string{"signal", "operation", "permission.created_at", "permission.updated_at"},
			wantVals: map[string]any{
				"operation": "validate",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := newMockLogger()
			tracer := newMockTracer()

			obs := NewPermissionObserver(logger, tracer)
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

func TestAdapter_Observer_PermissionObserver_OnSignal_WithError(t *testing.T) {
	tests := []struct {
		name    string
		sig     signal.SignalType
		data    signal.SignalPermission
		err     error
		wantMsg string
	}{
		{
			name: "permission with database error",
			sig:  signal.SignalFail,
			data: signal.SignalPermission{
				Operation: "create",
				UID:       util.Ptr("perm-123"),
			},
			err:     errors.New("database connection failed"),
			wantMsg: "service signal",
		},
		{
			name: "permission with validation error",
			sig:  signal.SignalReject,
			data: signal.SignalPermission{
				Operation: "update",
				Resource:  util.Ptr(""),
			},
			err:     errors.New("invalid resource"),
			wantMsg: "service signal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := newMockLogger()
			tracer := newMockTracer()

			obs := NewPermissionObserver(logger, tracer)
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

func TestAdapter_Observer_PermissionObserver_AllSignalTypes(t *testing.T) {
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

			obs := NewPermissionObserver(logger, tracer)
			ctx := context.Background()

			data := signal.SignalPermission{
				Operation: "test-operation",
				UID:       util.Ptr("test-perm"),
				Resource:  util.Ptr("test-resource"),
				Action:    util.Ptr("test-action"),
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

func TestAdapter_Observer_PermissionObserver_NilFields(t *testing.T) {
	logger := newMockLogger()
	tracer := newMockTracer()

	obs := NewPermissionObserver(logger, tracer)
	ctx := context.Background()

	// All nil fields except operation
	data := signal.SignalPermission{
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
	nilKeys := []string{"permission.uid", "permission.resource", "permission.action", "permission.description"}
	for _, key := range nilKeys {
		if _, ok := entry.fields[key]; ok {
			t.Errorf("key %q should not be present when nil", key)
		}
	}
}
