package observer

import (
	"context"
	"errors"
	"testing"

	"github.com/adityakw90/go-monitoring"
	"github.com/adityakw90/service-access/internal/core/domain/signal"
	"github.com/adityakw90/service-access/pkg/util"
)

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

func TestAdapter_Observer_NewAccessObserver(t *testing.T) {
	tests := []struct {
		name    string
		logger  *mockLogger
		tracer  monitoring.Tracer
		wantNil bool
	}{
		{
			name:    "create access observer with all parameters",
			logger:  newMockLogger(),
			tracer:  newMockTracer(),
			wantNil: false,
		},
		{
			name:    "create access observer with nil logger",
			logger:  nil,
			tracer:  newMockTracer(),
			wantNil: false,
		},
		{
			name:    "create access observer with nil tracer",
			logger:  newMockLogger(),
			tracer:  nil,
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obs := NewAccessObserver(tt.logger, tt.tracer)

			if (obs == nil) != tt.wantNil {
				t.Errorf("NewAccessObserver() = %v, wantNil %v", obs, tt.wantNil)
			}
		})
	}
}

func TestAdapter_Observer_AccessObserver_OnSignal_Success(t *testing.T) {
	tests := []struct {
		name     string
		sig      signal.SignalType
		data     signal.SignalAccessCheck
		wantKeys []string
		wantVals map[string]interface{}
	}{
		{
			name: "access check with all fields",
			sig:  signal.SignalStart,
			data: signal.SignalAccessCheck{
				SubjectID:   "user-123",
				SubjectType: "user",
				Resource:    "documents",
				Action:      "read",
			},
			wantKeys: []string{"signal", "subject_id", "subject_type", "resource", "action"},
			wantVals: map[string]interface{}{
				"subject_id":   "user-123",
				"subject_type": "user",
				"resource":     "documents",
				"action":       "read",
			},
		},
		{
			name: "access check with allowed true",
			sig:  signal.SignalSuccess,
			data: signal.SignalAccessCheck{
				SubjectID:   "user-456",
				SubjectType: "service",
				Resource:    "api",
				Action:      "write",
				Allowed:     util.Ptr(true),
			},
			wantKeys: []string{"signal", "subject_id", "subject_type", "resource", "action", "allowed"},
			wantVals: map[string]interface{}{
				"subject_id":   "user-456",
				"subject_type": "service",
				"resource":     "api",
				"action":       "write",
				"allowed":      true,
			},
		},
		{
			name: "access check with allowed false and reason",
			sig:  signal.SignalReject,
			data: signal.SignalAccessCheck{
				SubjectID:   "user-789",
				SubjectType: "user",
				Resource:    "admin",
				Action:      "delete",
				Allowed:     util.Ptr(false),
				Reason:      util.Ptr("insufficient permissions"),
			},
			wantKeys: []string{"signal", "subject_id", "subject_type", "resource", "action", "allowed", "reason"},
			wantVals: map[string]interface{}{
				"subject_id":   "user-789",
				"subject_type": "user",
				"resource":     "admin",
				"action":       "delete",
				"allowed":      false,
				"reason":       "insufficient permissions",
			},
		},
		{
			name: "access check with only reason",
			sig:  signal.SignalFail,
			data: signal.SignalAccessCheck{
				SubjectID:   "service-001",
				SubjectType: "service",
				Resource:    "metrics",
				Action:      "collect",
				Reason:      util.Ptr("rate limit exceeded"),
			},
			wantKeys: []string{"signal", "subject_id", "subject_type", "resource", "action", "reason"},
			wantVals: map[string]interface{}{
				"subject_id":   "service-001",
				"subject_type": "service",
				"resource":     "metrics",
				"action":       "collect",
				"reason":       "rate limit exceeded",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := newMockLogger()
			tracer := newMockTracer()

			obs := NewAccessObserver(logger, tracer)
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

func TestAdapter_Observer_AccessObserver_OnSignal_WithError(t *testing.T) {
	tests := []struct {
		name    string
		sig     signal.SignalType
		data    signal.SignalAccessCheck
		err     error
		wantMsg string
	}{
		{
			name: "access check with error",
			sig:  signal.SignalFail,
			data: signal.SignalAccessCheck{
				SubjectID:   "user-123",
				SubjectType: "user",
				Resource:    "documents",
				Action:      "read",
			},
			err:     errors.New("database connection failed"),
			wantMsg: "service signal",
		},
		{
			name: "access check with validation error",
			sig:  signal.SignalReject,
			data: signal.SignalAccessCheck{
				SubjectID:   "",
				SubjectType: "user",
				Resource:    "",
				Action:      "",
			},
			err:     errors.New("invalid subject ID"),
			wantMsg: "service signal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := newMockLogger()
			tracer := newMockTracer()

			obs := NewAccessObserver(logger, tracer)
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

func TestAdapter_Observer_AccessObserver_AllSignalTypes(t *testing.T) {
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

			obs := NewAccessObserver(logger, tracer)
			ctx := context.Background()

			data := signal.SignalAccessCheck{
				SubjectID:   "test-user",
				SubjectType: "user",
				Resource:    "test-resource",
				Action:      "test-action",
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
