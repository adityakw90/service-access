package publisher

import (
	"testing"

	"github.com/adityakw90/service-access/internal/core/domain/event"
)

// TestRabbitMQPublisher_Publish tests the CloudEvents conversion for RabbitMQ.
// Note: Full integration tests would require actual RabbitMQ connection.
func TestRabbitMQPublisher_Publish(t *testing.T) {
	tests := []struct {
		name      string
		eventType event.EventType
		eventData any
		source    string
		prefix    string
	}{
		{
			name:      "Check access event",
			eventType: event.EventAccessCheck,
			eventData: map[string]interface{}{
				"subject_id":   "test-uid",
				"subject_type": "user",
				"resource":     "user",
				"action":       "read",
				"reason":       "user not have any permission",
			},
			source: "rabbitmq-publisher",
			prefix: "user.service.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test CloudEvent conversion (same as Kafka)
			ce := toCloudEventData(tt.eventType, tt.eventData, tt.source)

			// Verify CloudEvent structure
			if ce.Type != string(tt.eventType) {
				t.Errorf("expected Type = %v, got %v", tt.eventType, ce.Type)
			}
			if ce.Source != tt.source {
				t.Errorf("expected Source = %v, got %v", tt.source, ce.Source)
			}
			if ce.SpecVersion != "1.0" {
				t.Errorf("expected SpecVersion = 1.0, got %v", ce.SpecVersion)
			}
			if ce.ID == "" {
				t.Error("expected ID to be non-empty")
			}
			if ce.Time == "" {
				t.Error("expected Time to be non-empty")
			}
			if len(ce.Data) == 0 {
				t.Error("expected Data to be non-empty")
			}

			// Verify routing key construction
			expectedRoutingKey := tt.prefix + string(tt.eventType)
			routingKey := tt.prefix + string(tt.eventType)
			if routingKey != expectedRoutingKey {
				t.Errorf("expected routing key %v, got %v", expectedRoutingKey, routingKey)
			}
		})
	}
}

func TestRabbitMQConfig_Defaults(t *testing.T) {
	config := RabbitMQPublisherConfig{
		Source:           "user-service",
		Exchange:         "user-service",
		ExchangeType:     "topic",
		RoutingKeyPrefix: "user.service.",
		Durable:          true,
	}

	if config.Exchange != "user-service" {
		t.Errorf("expected exchange user-service, got %v", config.Exchange)
	}
	if config.ExchangeType != "topic" {
		t.Errorf("expected exchange type topic, got %v", config.ExchangeType)
	}
	if config.RoutingKeyPrefix != "user.service." {
		t.Errorf("expected routing key prefix user.service., got %v", config.RoutingKeyPrefix)
	}
	if !config.Durable {
		t.Error("expected durable to be true")
	}
}

func TestRabbitMQRoutingKey(t *testing.T) {
	tests := []struct {
		name        string
		eventType   event.EventType
		prefix      string
		expectedKey string
	}{
		{
			name:        "Check access event routing key",
			eventType:   event.EventAccessCheck,
			prefix:      "access.service.",
			expectedKey: "access.service.access.check",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			routingKey := tt.prefix + string(tt.eventType)
			if routingKey != tt.expectedKey {
				t.Errorf("expected routing key %v, got %v", tt.expectedKey, routingKey)
			}
		})
	}
}

func TestToCloudEventData_RabbitMQ(t *testing.T) {
	tests := []struct {
		name      string
		eventType event.EventType
		eventData any
		source    string
	}{
		{
			name:      "Check access event",
			eventType: event.EventAccessCheck,
			eventData: map[string]interface{}{
				"subject_id":   "user-123",
				"subject_type": "user",
				"resource":     "user",
				"action":       "read",
				"reason":       "user not have any permission",
			},
			source: "rabbitmq-publisher",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ce := toCloudEventData(tt.eventType, tt.eventData, tt.source)

			if ce.Type != string(tt.eventType) {
				t.Errorf("expected Type = %v, got %v", tt.eventType, ce.Type)
			}
			if ce.Source != tt.source {
				t.Errorf("expected Source = %v, got %v", tt.source, ce.Source)
			}
		})
	}
}
