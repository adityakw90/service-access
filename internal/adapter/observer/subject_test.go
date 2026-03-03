package observer

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/adityakw90/go-monitoring"
	"github.com/adityakw90/service-access/internal/core/domain/signal"
)

func TestAdapter_Observer_NewSubjectObserver(t *testing.T) {
	tests := []struct {
		name    string
		logger  *mockLogger
		tracer  monitoring.Tracer
		wantNil bool
	}{
		{
			name:    "create subject observer with all parameters",
			logger:  newMockLogger(),
			tracer:  newMockTracer(),
			wantNil: false,
		},
		{
			name:    "create subject observer with nil logger",
			logger:  nil,
			tracer:  newMockTracer(),
			wantNil: false,
		},
		{
			name:    "create subject observer with nil tracer",
			logger:  newMockLogger(),
			tracer:  nil,
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obs := NewSubjectObserver(tt.logger, tt.tracer)

			if (obs == nil) != tt.wantNil {
				t.Errorf("NewSubjectObserver() = %v, wantNil %v", obs, tt.wantNil)
			}
		})
	}
}

func TestAdapter_Observer_SubjectObserver_OnSignal_Success(t *testing.T) {
	now := time.Now()
	past := now.Add(-24 * time.Hour)

	tests := []struct {
		name     string
		sig      signal.SignalType
		data     signal.SignalSubject
		wantKeys []string
		wantVals map[string]any
	}{
		{
			name: "subject signal with operation only",
			sig:  signal.SignalStart,
			data: signal.SignalSubject{
				Operation: "assign",
			},
			wantKeys: []string{"signal", "operation"},
			wantVals: map[string]any{
				"operation": "assign",
			},
		},
		{
			name: "subject signal with subject_id and type",
			sig:  signal.SignalSuccess,
			data: signal.SignalSubject{
				Operation:  "revoke",
				SubjectID:  stringPtr("user-123"),
				SubjectType: stringPtr("user"),
			},
			wantKeys: []string{"signal", "operation", "subject_id", "subject_type"},
			wantVals: map[string]any{
				"operation":    "revoke",
				"subject_id":   "user-123",
				"subject_type": "user",
			},
		},
		{
			name: "subject signal with role_uid",
			sig:  signal.SignalStart,
			data: signal.SignalSubject{
				Operation:  "assign",
				SubjectID:  stringPtr("service-456"),
				SubjectType: stringPtr("service"),
				RoleUID:     stringPtr("role-admin"),
			},
			wantKeys: []string{"signal", "operation", "subject_id", "subject_type", "role_uid"},
			wantVals: map[string]any{
				"operation":    "assign",
				"subject_id":   "service-456",
				"subject_type": "service",
				"role_uid":     "role-admin",
			},
		},
		{
			name: "subject signal with all fields",
			sig:  signal.SignalReject,
			data: signal.SignalSubject{
				Operation:   "unassign",
				SubjectID:   stringPtr("user-789"),
				SubjectType: stringPtr("user"),
				RoleUID:     stringPtr("role-editor"),
				AssignedAt:  &now,
			},
			wantKeys: []string{"signal", "operation", "subject_id", "subject_type", "role_uid", "assigned_at"},
			wantVals: map[string]any{
				"operation":    "unassign",
				"subject_id":   "user-789",
				"subject_type": "user",
				"role_uid":     "role-editor",
			},
		},
		{
			name: "subject signal with timestamp",
			sig:  signal.SignalFail,
			data: signal.SignalSubject{
				Operation:  "validate",
				AssignedAt: &past,
			},
			wantKeys: []string{"signal", "operation", "assigned_at"},
			wantVals: map[string]any{
				"operation": "validate",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := newMockLogger()
			tracer := newMockTracer()

			obs := NewSubjectObserver(logger, tracer)
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

func TestAdapter_Observer_SubjectObserver_OnSignal_WithError(t *testing.T) {
	tests := []struct {
		name    string
		sig     signal.SignalType
		data    signal.SignalSubject
		err     error
		wantMsg string
	}{
		{
			name: "subject with database error",
			sig:  signal.SignalFail,
			data: signal.SignalSubject{
				Operation: "assign",
				SubjectID: stringPtr("user-123"),
				RoleUID:   stringPtr("role-admin"),
			},
			err:     errors.New("database connection failed"),
			wantMsg: "service signal",
		},
		{
			name: "subject with validation error",
			sig:  signal.SignalReject,
			data: signal.SignalSubject{
				Operation: "revoke",
				SubjectID: stringPtr(""),
			},
			err:     errors.New("invalid subject ID"),
			wantMsg: "service signal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := newMockLogger()
			tracer := newMockTracer()

			obs := NewSubjectObserver(logger, tracer)
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

func TestAdapter_Observer_SubjectObserver_AllSignalTypes(t *testing.T) {
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

			obs := NewSubjectObserver(logger, tracer)
			ctx := context.Background()

			data := signal.SignalSubject{
				Operation:   "test-operation",
				SubjectID:   stringPtr("test-subject"),
				SubjectType: stringPtr("user"),
				RoleUID:     stringPtr("test-role"),
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

func TestAdapter_Observer_SubjectObserver_NilFields(t *testing.T) {
	logger := newMockLogger()
	tracer := newMockTracer()

	obs := NewSubjectObserver(logger, tracer)
	ctx := context.Background()

	// All nil fields except operation
	data := signal.SignalSubject{
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
	nilKeys := []string{"subject_id", "subject_type", "role_uid", "assigned_at"}
	for _, key := range nilKeys {
		if _, ok := entry.fields[key]; ok {
			t.Errorf("key %q should not be present when nil", key)
		}
	}
}

func TestAdapter_Observer_SubjectObserver_DifferentSubjectTypes(t *testing.T) {
	tests := []struct {
		name        string
		subjectType string
		subjectID   string
	}{
		{
			name:        "user subject",
			subjectType: "user",
			subjectID:   "user-123",
		},
		{
			name:        "service subject",
			subjectType: "service",
			subjectID:   "service-api",
		},
		{
			name:        "api_key subject",
			subjectType: "api_key",
			subjectID:   "key-abc123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := newMockLogger()
			tracer := newMockTracer()

			obs := NewSubjectObserver(logger, tracer)
			ctx := context.Background()

			data := signal.SignalSubject{
				Operation:   "assign",
				SubjectID:   &tt.subjectID,
				SubjectType: &tt.subjectType,
				RoleUID:     stringPtr("role-test"),
			}
			obs.OnSignal(ctx, signal.SignalStart, data, nil)

			if len(logger.debugMessages) != 1 {
				t.Fatalf("expected 1 debug log entry, got %d", len(logger.debugMessages))
			}

			entry := logger.debugMessages[0]
			if entry.fields["subject_type"] != tt.subjectType {
				t.Errorf("subject_type = %v, want %v", entry.fields["subject_type"], tt.subjectType)
			}
			if entry.fields["subject_id"] != tt.subjectID {
				t.Errorf("subject_id = %v, want %v", entry.fields["subject_id"], tt.subjectID)
			}
		})
	}
}
