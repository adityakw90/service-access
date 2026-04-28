package event

import (
	"context"
	"encoding/json"
	"regexp"
	"testing"
	"time"

	"github.com/adityakw90/service-access/internal/core/domain/event"
	"github.com/adityakw90/service-access/pkg/util"
)

func isValidUUID(id string) bool {
	pattern := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	return pattern.MatchString(id)
}

func isValidRFC3339(timestamp string) bool {
	_, err := time.Parse(time.RFC3339, timestamp)
	return err == nil
}

func setupContext(ctx context.Context, clientName, actorId, actorType string) context.Context {
	ctx = util.SetClientName(ctx, clientName)
	ctx = util.SetActor(ctx, actorId, actorType)
	return ctx
}

func TestNewCloudEvent(t *testing.T) {
	tests := []struct {
		name        string
		setupCtx    func() context.Context
		eventType   event.EventType
		eventData   any
		wantSource  string
		wantSpecVer string
		validate    func(t *testing.T, ce CloudEvent)
	}{
		{
			name: "Happy Path - EventAccessCheckData with full context",
			setupCtx: func() context.Context {
				return setupContext(context.Background(), "test-client", "user-123", "user")
			},
			eventType: event.EventAccessCheck,
			eventData: event.EventAccessCheckData{
				SubjectId:   "user-123",
				SubjectType: "user",
				Resource:    "document:123",
				Action:      "read",
				Reason:      "permission granted",
			},
			wantSource:  Source,
			wantSpecVer: SpecVersion,
			validate: func(t *testing.T, ce CloudEvent) {
				if ce.Type != string(event.EventAccessCheck) {
					t.Errorf("Type = %v, want %v", ce.Type, event.EventAccessCheck)
				}
				if ce.Data.Client != "test-client" {
					t.Errorf("Client = %v, want %v", ce.Data.Client, "test-client")
				}
				if ce.Data.ActorId != "user-123" {
					t.Errorf("ActorId = %v, want %v", ce.Data.ActorId, "user-123")
				}
				if ce.Data.ActorType != "user" {
					t.Errorf("ActorType = %v, want %v", ce.Data.ActorType, "user")
				}
				if !isValidUUID(ce.ID) {
					t.Errorf("ID = %v is not a valid UUID", ce.ID)
				}
				if !isValidRFC3339(ce.Time) {
					t.Errorf("Time = %v is not valid RFC3339", ce.Time)
				}
				var accessCheckData event.EventAccessCheckData
				if err := json.Unmarshal(ce.Data.MetaData, &accessCheckData); err != nil {
					t.Errorf("Failed to unmarshal MetaData: %v", err)
				}
				if accessCheckData.Resource != "document:123" {
					t.Errorf("Resource = %v, want %v", accessCheckData.Resource, "document:123")
				}
			},
		},
		{
			name: "Context Defaults - Empty context uses unknown values",
			setupCtx: func() context.Context {
				return context.Background()
			},
			eventType: event.EventAccessCheck,
			eventData: event.EventAccessCheckData{Resource: "file:456", Action: "write"},
			wantSource:  Source,
			wantSpecVer: SpecVersion,
			validate: func(t *testing.T, ce CloudEvent) {
				if ce.Data.Client != "unknown" {
					t.Errorf("Client = %v, want %v", ce.Data.Client, "unknown")
				}
				if ce.Data.ActorId != "unknown" {
					t.Errorf("ActorId = %v, want %v", ce.Data.ActorId, "unknown")
				}
				if ce.Data.ActorType != "unknown" {
					t.Errorf("ActorType = %v, want %v", ce.Data.ActorType, "unknown")
				}
			},
		},
		{
			name: "EventRoleCreateData",
			setupCtx: func() context.Context {
				return setupContext(context.Background(), "admin-panel", "admin-456", "admin")
			},
			eventType: event.EventRoleCreate,
			eventData: event.EventRoleCreateData{
				UID:         "role-uid-123",
				GroupUID:    "group-uid-123",
				Name:        "Editor",
				Description: "Can edit documents",
				CreatedAt:   time.Now(),
			},
			wantSource:  Source,
			wantSpecVer: SpecVersion,
			validate: func(t *testing.T, ce CloudEvent) {
				if ce.Type != string(event.EventRoleCreate) {
					t.Errorf("Type = %v, want %v", ce.Type, event.EventRoleCreate)
				}
				var data event.EventRoleCreateData
				if err := json.Unmarshal(ce.Data.MetaData, &data); err != nil {
					t.Errorf("Failed to unmarshal MetaData: %v", err)
				}
				if data.Name != "Editor" {
					t.Errorf("Name = %v, want %v", data.Name, "Editor")
				}
			},
		},
		{
			name: "EventPermissionCreateData",
			setupCtx: func() context.Context {
				return setupContext(context.Background(), "admin-panel", "admin-789", "admin")
			},
			eventType: event.EventPermissionCreate,
			eventData: event.EventPermissionCreateData{
				UID:         "perm-uid-789",
				Resource:    "document",
				Action:      "delete",
				Description: "Can delete documents",
				CreatedAt:   time.Now(),
			},
			wantSource:  Source,
			wantSpecVer: SpecVersion,
			validate: func(t *testing.T, ce CloudEvent) {
				if ce.Type != string(event.EventPermissionCreate) {
					t.Errorf("Type = %v, want %v", ce.Type, event.EventPermissionCreate)
				}
				var data event.EventPermissionCreateData
				if err := json.Unmarshal(ce.Data.MetaData, &data); err != nil {
					t.Errorf("Failed to unmarshal MetaData: %v", err)
				}
				if data.Resource != "document" {
					t.Errorf("Resource = %v, want %v", data.Resource, "document")
				}
			},
		},
		{
			name: "EventSubjectAssignData",
			setupCtx: func() context.Context {
				return setupContext(context.Background(), "api", "system-1", "system")
			},
			eventType: event.EventSubjectAssign,
			eventData: event.EventSubjectAssignData{
				SubjectID:   "user-subject-123",
				SubjectType: "user",
				RoleUID:     "role-uid-123",
				AssignedAt:  time.Now(),
			},
			wantSource:  Source,
			wantSpecVer: SpecVersion,
			validate: func(t *testing.T, ce CloudEvent) {
				if ce.Type != string(event.EventSubjectAssign) {
					t.Errorf("Type = %v, want %v", ce.Type, event.EventSubjectAssign)
				}
				var data event.EventSubjectAssignData
				if err := json.Unmarshal(ce.Data.MetaData, &data); err != nil {
					t.Errorf("Failed to unmarshal MetaData: %v", err)
				}
				if data.SubjectID != "user-subject-123" {
					t.Errorf("SubjectID = %v, want %v", data.SubjectID, "user-subject-123")
				}
			},
		},
		{
			name: "Marshal Error - Non-serializable data",
			setupCtx: func() context.Context {
				return setupContext(context.Background(), "test", "actor-1", "system")
			},
			eventType:   event.EventAccessCheck,
			eventData:   make(chan int),
			wantSource:  Source,
			wantSpecVer: SpecVersion,
			validate: func(t *testing.T, ce CloudEvent) {
				var errData map[string]any
				if err := json.Unmarshal(ce.Data.MetaData, &errData); err != nil {
					t.Fatalf("Failed to unmarshal error data: %v", err)
				}
				if _, ok := errData["error"]; !ok {
					t.Errorf("MetaData should contain 'error' key, got %v", errData)
				}
			},
		},
		{
			name: "Group Created Event",
			setupCtx: func() context.Context {
				return setupContext(context.Background(), "admin-panel", "admin-1", "admin")
			},
			eventType: event.EventGroupCreate,
			eventData: struct {
				UID         string `json:"uid"`
				Name        string `json:"name"`
				Description string `json:"description"`
			}{
				UID:         "group-uid-123",
				Name:        "Editors",
				Description: "Users who can edit",
			},
			wantSource:  Source,
			wantSpecVer: SpecVersion,
			validate: func(t *testing.T, ce CloudEvent) {
				if ce.Type != string(event.EventGroupCreate) {
					t.Errorf("Type = %v, want %v", ce.Type, event.EventGroupCreate)
				}
				if ce.Data.Client != "admin-panel" {
					t.Errorf("Client = %v, want %v", ce.Data.Client, "admin-panel")
				}
			},
		},
		{
			name: "Permission Update Event",
			setupCtx: func() context.Context {
				return setupContext(context.Background(), "admin-panel", "admin-2", "admin")
			},
			eventType: event.EventPermissionUpdate,
			eventData: struct {
				UID         string `json:"uid"`
				Resource    string `json:"resource"`
				Action      string `json:"action"`
				Description string `json:"description"`
			}{
				UID:         "perm-uid-456",
				Resource:    "report",
				Action:      "generate",
				Description: "Can generate reports",
			},
			wantSource:  Source,
			wantSpecVer: SpecVersion,
			validate: func(t *testing.T, ce CloudEvent) {
				if ce.Type != string(event.EventPermissionUpdate) {
					t.Errorf("Type = %v, want %v", ce.Type, event.EventPermissionUpdate)
				}
				var data event.EventPermissionUpdateData
				if err := json.Unmarshal(ce.Data.MetaData, &data); err != nil {
					t.Errorf("Failed to unmarshal MetaData: %v", err)
				}
				if data.Resource != "report" {
					t.Errorf("Resource = %v, want %v", data.Resource, "report")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := tt.setupCtx()
			ce := NewCloudEvent(ctx, tt.eventType, tt.eventData)

			if ce.Source != tt.wantSource {
				t.Errorf("Source = %v, want %v", ce.Source, tt.wantSource)
			}
			if ce.SpecVersion != tt.wantSpecVer {
				t.Errorf("SpecVersion = %v, want %v", ce.SpecVersion, tt.wantSpecVer)
			}
			if tt.validate != nil {
				tt.validate(t, ce)
			}
		})
	}
}

func TestNewCloudEvent_Constants(t *testing.T) {
	t.Parallel()
	ctx := setupContext(context.Background(), "test", "actor-1", "system")
	ce := NewCloudEvent(ctx, event.EventAccessCheck, event.EventAccessCheckData{})

	if ce.Source != Source {
		t.Errorf("Source = %v, want %v", ce.Source, Source)
	}
	if ce.SpecVersion != SpecVersion {
		t.Errorf("SpecVersion = %v, want %v", ce.SpecVersion, SpecVersion)
	}
}

func TestNewCloudEvent_IDGeneration(t *testing.T) {
	t.Parallel()
	ctx := setupContext(context.Background(), "test", "actor-1", "system")

	ids := make(map[string]bool)
	for range 100 {
		ce := NewCloudEvent(ctx, event.EventAccessCheck, event.EventAccessCheckData{})
		if !isValidUUID(ce.ID) {
			t.Errorf("ID = %v is not a valid UUID", ce.ID)
		}
		if ids[ce.ID] {
			t.Errorf("Duplicate ID generated: %v", ce.ID)
		}
		ids[ce.ID] = true
	}
}

func TestNewCloudEvent_TimeFormat(t *testing.T) {
	t.Parallel()
	ctx := setupContext(context.Background(), "test", "actor-1", "system")

	ce := NewCloudEvent(ctx, event.EventAccessCheck, event.EventAccessCheckData{})

	parsedTime, err := time.Parse(time.RFC3339, ce.Time)
	if err != nil {
		t.Fatalf("Failed to parse Time: %v", err)
	}

	now := time.Now().UTC()
	if parsedTime.After(now.Add(5 * time.Second)) {
		t.Errorf("Time = %v is too far in the future", ce.Time)
	}
	if parsedTime.Before(now.Add(-5 * time.Second)) {
		t.Errorf("Time = %v is too far in the past", ce.Time)
	}
}
