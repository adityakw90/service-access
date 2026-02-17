package publisher

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/adityakw90/service-access/internal/core/domain/event"
)

func TestToCloudEventData(t *testing.T) {
	tests := []struct {
		name      string
		eventType event.EventType
		eventData any
		source    string
		wantType  string
	}{
		{
			name:      "Login event with simple data",
			eventType: event.EventAccessCheck,
			eventData: event.EventAccessCheckData{
				SubjectId:   "uid-from-user-service",
				SubjectType: "user",
				Resource:    "user",
				Action:      "read",
				Reason:      "user not have any permission",
			},
			source:   "service-access",
			wantType: "access.check",
		},
		{
			name:      "Failed login event with simple data",
			eventType: event.EventAccessCheckFailed,
			eventData: event.EventAccessCheckFailedData{
				SubjectId:   "uid-from-user-service",
				SubjectType: "user",
				Resource:    "user",
				Action:      "read",
				Reason:      "user not have any permission",
			},
			source:   "service-access",
			wantType: "access.check_failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toCloudEventData(tt.eventType, tt.eventData, tt.source)

			fmt.Println(got)

			if got.Type != tt.wantType {
				t.Errorf("toCloudEventData() Type = %v, want %v", got.Type, tt.wantType)
			}

			if got.Source != tt.source {
				t.Errorf("toCloudEventData() Source = %v, want %v", got.Source, tt.source)
			}

			if got.SpecVersion != "1.0" {
				t.Errorf("toCloudEventData() SpecVersion = %v, want %v", got.SpecVersion, "1.0")
			}

			if got.ID == "" {
				t.Errorf("toCloudEventData() ID should not be empty")
			}

			if got.Time == "" {
				t.Errorf("toCloudEventData() Time should not be empty")
			}

			// Verify data can be unmarshaled
			var data map[string]interface{}
			if err := json.Unmarshal(got.Data, &data); err != nil {
				t.Errorf("toCloudEventData() Data is not valid JSON: %v", err)
			}
		})
	}
}

func TestToCloudEvent(t *testing.T) {
	tests := []struct {
		name     string
		event    any
		source   string
		wantType string
		wantErr  bool
	}{
		{
			name: "Check access event data",
			event: event.EventAccessCheckData{
				SubjectId:   "uid-from-user-service",
				SubjectType: "user",
				Resource:    "user",
				Action:      "read",
				Reason:      "user not have any permission",
			},
			source:   "service-access",
			wantType: "",
			wantErr:  false,
		},
		{
			name: "Check access failed event data",
			event: event.EventAccessCheckFailedData{
				SubjectId:   "uid-from-user-service",
				SubjectType: "user",
				Resource:    "user",
				Action:      "read",
				Reason:      "user not have any permission",
			},
			source:   "service-access",
			wantType: "",
			wantErr:  false,
		},
		{
			name: "Simple map",
			event: map[string]any{
				"user_uid": "user-123",
				"email":    "test@example.com",
			},
			source:   "service-access",
			wantType: "",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToCloudEvent(tt.event, tt.source)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToCloudEvent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got.Source != tt.source {
				t.Errorf("ToCloudEvent() Source = %v, want %v", got.Source, tt.source)
			}

			if got.SpecVersion != "1.0" {
				t.Errorf("ToCloudEvent() SpecVersion = %v, want %v", got.SpecVersion, "1.0")
			}

			if got.ID == "" {
				t.Errorf("ToCloudEvent() ID should not be empty")
			}

			if got.Time == "" {
				t.Errorf("ToCloudEvent() Time should not be empty")
			}

			// Verify data can be unmarshaled
			var data map[string]interface{}
			if err := json.Unmarshal(got.Data, &data); err != nil {
				t.Errorf("ToCloudEvent() Data is not valid JSON: %v", err)
			}
		})
	}
}
