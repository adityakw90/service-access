package event

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/adityakw90/service-access/internal/core/domain/event"
	"github.com/stretchr/testify/assert"
)

type mockRabbitConn struct {
	publishWithConfirmCalled            bool
	publishWithConfirmArgCtx            context.Context
	publishWithConfirmArgExchange       string
	publishWithConfirmArgRoutingKey     string
	publishWithConfirmArgContentType    string
	publishWithConfirmArgHeaders        map[string]string
	publishWithConfirmArgBody           []byte
	publishWithConfirmArgDeliveryMode   *uint8
	publishWithConfirmArgPriority       *uint8
	publishWithConfirmArgTimestamp      *time.Time
	publishWithConfirmArgExpiration     *time.Duration
	publishWithConfirmArgMaxRetries     int
	publishWithConfirmArgRetryInterval  time.Duration
	publishWithConfirmArgConfirmTimeout time.Duration
	publishWithConfirmReturnErr         error
	publishWithConfirmCheckContext      bool

	closeCalled    bool
	closeReturnErr error
}

func (m *mockRabbitConn) PublishWithConfirm(
	ctx context.Context,
	exchange string,
	routingKey string,
	contentType string,
	headers map[string]string,
	body []byte,
	deliveryMode *uint8,
	priority *uint8,
	timestamp *time.Time,
	expiration *time.Duration,
	maxRetries int,
	retryInterval time.Duration,
	confirmTimeout time.Duration,
) error {
	if m.publishWithConfirmCheckContext && ctx.Err() != nil {
		return ctx.Err()
	}
	m.publishWithConfirmCalled = true
	m.publishWithConfirmArgCtx = ctx
	m.publishWithConfirmArgExchange = exchange
	m.publishWithConfirmArgRoutingKey = routingKey
	m.publishWithConfirmArgContentType = contentType
	m.publishWithConfirmArgHeaders = headers
	m.publishWithConfirmArgBody = body
	m.publishWithConfirmArgDeliveryMode = deliveryMode
	m.publishWithConfirmArgPriority = priority
	m.publishWithConfirmArgTimestamp = timestamp
	m.publishWithConfirmArgExpiration = expiration
	m.publishWithConfirmArgMaxRetries = maxRetries
	m.publishWithConfirmArgRetryInterval = retryInterval
	m.publishWithConfirmArgConfirmTimeout = confirmTimeout
	return m.publishWithConfirmReturnErr
}

func (m *mockRabbitConn) Close() error {
	m.closeCalled = true
	return m.closeReturnErr
}

func TestNewRabbitmqPublisher(t *testing.T) {
	mockConn := &mockRabbitConn{}
	config := RabbitmqPublisherConfig{
		Exchange:         "test-exchange",
		RoutingKeyPrefix: "user.service",
		ConfirmTimeout:   5 * time.Second,
		MaxRetries:       3,
		RetryInterval:    500 * time.Millisecond,
	}
	logger := newMockLogger()
	tracer := newMockTracer()

	publisher := NewRabbitmqPublisher(mockConn, config, logger, tracer)

	rabbitPub, ok := publisher.(*RabbitmqPublisher)
	assert.True(t, ok, "Publisher should be *RabbitmqPublisher type")
	assert.Equal(t, mockConn, rabbitPub.conn)
	assert.Equal(t, "test-exchange", rabbitPub.exchange)
	assert.Equal(t, "user.service.", rabbitPub.routingKeyPrefix)
	assert.Equal(t, 5*time.Second, rabbitPub.confirmTimeout)
	assert.Equal(t, 3, rabbitPub.maxRetries)
	assert.Equal(t, 500*time.Millisecond, rabbitPub.retryInterval)
	assert.Equal(t, logger, rabbitPub.logger)
	assert.Equal(t, tracer, rabbitPub.tracer)
}

func TestRabbitmqPublisher_Name(t *testing.T) {
	mockConn := &mockRabbitConn{}
	config := RabbitmqPublisherConfig{}
	logger := newMockLogger()
	tracer := newMockTracer()

	publisher := NewRabbitmqPublisher(mockConn, config, logger, tracer)

	assert.Equal(t, "RabbitmqPublisher", publisher.Name())
}

func TestRabbitmqPublisher_Publish(t *testing.T) {
	tests := []struct {
		name        string
		setupCtx    func() context.Context
		eventType   event.EventType
		eventData   any
		config      RabbitmqPublisherConfig
		setupMock   func(*mockRabbitConn)
		wantErr     bool
		errContains string
		validate    func(*testing.T, *mockRabbitConn)
	}{
		{
			name: "Happy Path - EventAccessCheck with full context",
			setupCtx: func() context.Context {
				return setupContext(context.Background(), "test-client", "user-123", "user")
			},
			eventType: event.EventAccessCheck,
			eventData: event.EventAccessCheckData{
				SubjectId:   "user-123",
				SubjectType: "user",
				Resource:    "document:456",
				Action:      "read",
				Reason:      "allowed",
			},
			config: RabbitmqPublisherConfig{
				Exchange:         "test-exchange",
				RoutingKeyPrefix: "user.service",
				ConfirmTimeout:   5 * time.Second,
				MaxRetries:       3,
				RetryInterval:    500 * time.Millisecond,
			},
			setupMock: func(m *mockRabbitConn) {
				m.publishWithConfirmReturnErr = nil
			},
			wantErr: false,
			validate: func(t *testing.T, m *mockRabbitConn) {
				assert.True(t, m.publishWithConfirmCalled, "PublishWithConfirm should be called")
				assert.Equal(t, "test-exchange", m.publishWithConfirmArgExchange)
				assert.Equal(t, "user.service.access.check", m.publishWithConfirmArgRoutingKey)
				assert.Equal(t, "application/cloudevents+json", m.publishWithConfirmArgContentType)

				assert.NotNil(t, m.publishWithConfirmArgHeaders)
				assert.NotEmpty(t, m.publishWithConfirmArgHeaders["ce_type"])
				assert.NotEmpty(t, m.publishWithConfirmArgHeaders["ce_source"])
				assert.NotEmpty(t, m.publishWithConfirmArgHeaders["ce_id"])
				assert.NotEmpty(t, m.publishWithConfirmArgHeaders["ce_specversion"])
				assert.Equal(t, "access.check", m.publishWithConfirmArgHeaders["ce_type"])
				assert.Equal(t, Source, m.publishWithConfirmArgHeaders["ce_source"])
				assert.Equal(t, SpecVersion, m.publishWithConfirmArgHeaders["ce_specversion"])

				assert.NotNil(t, m.publishWithConfirmArgBody)
				assert.NotEmpty(t, m.publishWithConfirmArgBody)

				assert.Equal(t, uint8(2), *m.publishWithConfirmArgDeliveryMode)
				assert.Equal(t, uint8(10), *m.publishWithConfirmArgPriority)
				assert.Nil(t, m.publishWithConfirmArgTimestamp)
				assert.Nil(t, m.publishWithConfirmArgExpiration)
				assert.Equal(t, 3, m.publishWithConfirmArgMaxRetries)
				assert.Equal(t, 500*time.Millisecond, m.publishWithConfirmArgRetryInterval)
				assert.Equal(t, 5*time.Second, m.publishWithConfirmArgConfirmTimeout)
			},
		},
		{
			name: "Happy Path - EventRoleCreate",
			setupCtx: func() context.Context {
				return setupContext(context.Background(), "admin-client", "admin-456", "admin")
			},
			eventType: event.EventRoleCreate,
			eventData: event.EventRoleCreateData{
				UID:         "role-uid-123",
				GroupUID:    "group-uid-123",
				Name:        "Editor",
				Description: "Can edit documents",
				CreatedAt:   time.Now(),
			},
			config: RabbitmqPublisherConfig{
				Exchange:         "role-events",
				RoutingKeyPrefix: "role",
			},
			setupMock: func(m *mockRabbitConn) {
				m.publishWithConfirmReturnErr = nil
			},
			wantErr: false,
			validate: func(t *testing.T, m *mockRabbitConn) {
				assert.True(t, m.publishWithConfirmCalled)
				assert.Equal(t, "role-events", m.publishWithConfirmArgExchange)
				assert.Equal(t, "role.role.create", m.publishWithConfirmArgRoutingKey)
				assert.Equal(t, "role.create", m.publishWithConfirmArgHeaders["ce_type"])
			},
		},
		{
			name: "Happy Path - Context without client/actor uses defaults",
			setupCtx: func() context.Context {
				return context.Background()
			},
			eventType: event.EventPermissionCreate,
			eventData: event.EventPermissionCreateData{
				UID:         "perm-uid-456",
				Resource:    "file",
				Action:      "write",
				Description: "Can write files",
				CreatedAt:   time.Now(),
			},
			config: RabbitmqPublisherConfig{
				Exchange:         "events",
				RoutingKeyPrefix: "",
			},
			setupMock: func(m *mockRabbitConn) {
				m.publishWithConfirmReturnErr = nil
			},
			wantErr: false,
			validate: func(t *testing.T, m *mockRabbitConn) {
				assert.True(t, m.publishWithConfirmCalled)
				assert.Equal(t, "permission.create", m.publishWithConfirmArgRoutingKey)
			},
		},
		{
			name: "Happy Path - EventGroupCreate",
			setupCtx: func() context.Context {
				return setupContext(context.Background(), "api", "admin-789", "admin")
			},
			eventType: event.EventGroupCreate,
			eventData: event.EventGroupCreateData{
				UID:         "group-uid-789",
				Name:        "Editors",
				Description: "Users who can edit",
				CreatedAt:   time.Now(),
			},
			config: RabbitmqPublisherConfig{
				Exchange:         "security-events",
				RoutingKeyPrefix: "access",
			},
			setupMock: func(m *mockRabbitConn) {
				m.publishWithConfirmReturnErr = nil
			},
			wantErr: false,
			validate: func(t *testing.T, m *mockRabbitConn) {
				assert.True(t, m.publishWithConfirmCalled)
				assert.Equal(t, "access.group.create", m.publishWithConfirmArgRoutingKey)
			},
		},
		{
			name: "Happy Path - EventSubjectAssign",
			setupCtx: func() context.Context {
				return setupContext(context.Background(), "mobile", "user-101", "user")
			},
			eventType: event.EventSubjectAssign,
			eventData: event.EventSubjectAssignData{
				SubjectID:   "user-101",
				SubjectType: "user",
				RoleUID:     "role-uid-101",
				AssignedAt:  time.Now(),
			},
			config: RabbitmqPublisherConfig{
				Exchange:         "events",
				RoutingKeyPrefix: "subject",
			},
			setupMock: func(m *mockRabbitConn) {
				m.publishWithConfirmReturnErr = nil
			},
			wantErr: false,
			validate: func(t *testing.T, m *mockRabbitConn) {
				assert.True(t, m.publishWithConfirmCalled)
				assert.Equal(t, "subject.subject.assign", m.publishWithConfirmArgRoutingKey)
			},
		},
		{
			name: "Happy Path - EventSubjectRevoke",
			setupCtx: func() context.Context {
				return setupContext(context.Background(), "web", "user-202", "user")
			},
			eventType: event.EventSubjectRevoke,
			eventData: event.EventSubjectRevokeData{
				SubjectID:   "user-202",
				SubjectType: "user",
				RoleUID:     "role-uid-202",
				RevokedAt:   time.Now(),
			},
			config: RabbitmqPublisherConfig{
				Exchange:         "events",
				RoutingKeyPrefix: "subject",
			},
			setupMock: func(m *mockRabbitConn) {
				m.publishWithConfirmReturnErr = nil
			},
			wantErr: false,
		},
		{
			name: "Happy Path - EventRoleUpdate",
			setupCtx: func() context.Context {
				return setupContext(context.Background(), "admin", "admin-303", "admin")
			},
			eventType: event.EventRoleUpdate,
			eventData: event.EventRoleUpdateData{
				UID:         "role-uid-303",
				Name:        "Moderator",
				Description: "Can moderate content",
				UpdatedAt:   time.Now(),
			},
			config: RabbitmqPublisherConfig{
				Exchange:         "events",
				RoutingKeyPrefix: "role",
			},
			setupMock: func(m *mockRabbitConn) {
				m.publishWithConfirmReturnErr = nil
			},
			wantErr: false,
		},
		{
			name: "Happy Path - EventPermissionDelete",
			setupCtx: func() context.Context {
				return setupContext(context.Background(), "api", "admin-404", "admin")
			},
			eventType: event.EventPermissionDelete,
			eventData: event.EventPermissionDeleteData{
				UID: "perm-uid-404",
			},
			config: RabbitmqPublisherConfig{
				Exchange:         "permission-events",
				RoutingKeyPrefix: "permission",
			},
			setupMock: func(m *mockRabbitConn) {
				m.publishWithConfirmReturnErr = nil
			},
			wantErr: false,
		},
		{
			name: "Happy Path - EventGroupDelete",
			setupCtx: func() context.Context {
				return setupContext(context.Background(), "api", "admin-505", "admin")
			},
			eventType: event.EventGroupDelete,
			eventData: event.EventGroupDeleteData{
				UID: "group-uid-505",
			},
			config: RabbitmqPublisherConfig{
				Exchange:         "group-events",
				RoutingKeyPrefix: "group",
			},
			setupMock: func(m *mockRabbitConn) {
				m.publishWithConfirmReturnErr = nil
			},
			wantErr: false,
		},
		{
			name: "Happy Path - EventRoleDelete",
			setupCtx: func() context.Context {
				return setupContext(context.Background(), "api", "admin-606", "admin")
			},
			eventType: event.EventRoleDelete,
			eventData: event.EventRoleDeleteData{
				UID: "role-uid-606",
			},
			config: RabbitmqPublisherConfig{
				Exchange:         "role-events",
				RoutingKeyPrefix: "role",
			},
			setupMock: func(m *mockRabbitConn) {
				m.publishWithConfirmReturnErr = nil
			},
			wantErr: false,
		},
		{
			name: "Error - RabbitMQ publish failure",
			setupCtx: func() context.Context {
				return setupContext(context.Background(), "test-client", "user-123", "user")
			},
			eventType: event.EventAccessCheck,
			eventData: event.EventAccessCheckData{
				SubjectId: "user-123",
				Resource:  "doc:789",
				Action:    "read",
			},
			config: RabbitmqPublisherConfig{},
			setupMock: func(m *mockRabbitConn) {
				m.publishWithConfirmReturnErr = errors.New("connection failed")
			},
			wantErr:     true,
			errContains: "failed to publish message to RabbitMQ",
		},
		{
			name: "Error - Context cancellation",
			setupCtx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			eventType: event.EventAccessCheck,
			eventData: event.EventAccessCheckData{
				SubjectId: "user-456",
				Resource:  "file:123",
				Action:    "write",
			},
			config: RabbitmqPublisherConfig{},
			setupMock: func(m *mockRabbitConn) {
				m.publishWithConfirmReturnErr = nil
				m.publishWithConfirmCheckContext = true
			},
			wantErr:     true,
			errContains: "failed to publish message to RabbitMQ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConn := &mockRabbitConn{}
			tt.setupMock(mockConn)

			logger := newMockLogger()
			tracer := newMockTracer()

			publisher := NewRabbitmqPublisher(mockConn, tt.config, logger, tracer)

			ctx := tt.setupCtx()
			err := publisher.Publish(ctx, tt.eventType, tt.eventData)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}

			if tt.validate != nil {
				tt.validate(t, mockConn)
			}
		})
	}
}

func TestRabbitmqPublisher_Close(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockRabbitConn)
		wantErr     bool
		errContains string
	}{
		{
			name: "Happy Path - Successful close",
			setupMock: func(m *mockRabbitConn) {
				m.closeReturnErr = nil
			},
			wantErr: false,
		},
		{
			name: "Error - Close propagates RabbitMQ error",
			setupMock: func(m *mockRabbitConn) {
				m.closeReturnErr = errors.New("connection already closed")
			},
			wantErr:     true,
			errContains: "connection already closed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConn := &mockRabbitConn{}
			tt.setupMock(mockConn)

			config := RabbitmqPublisherConfig{}
			logger := newMockLogger()
			tracer := newMockTracer()

			publisher := NewRabbitmqPublisher(mockConn, config, logger, tracer)

			err := publisher.Close()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}

			assert.True(t, mockConn.closeCalled, "Close should be called on underlying Rabbit connection")
		})
	}
}

func TestRabbitmqPublisher_getRoutingKey(t *testing.T) {
	tests := []struct {
		name            string
		prefix          string
		eventType       string
		expectedRouting string
	}{
		{
			name:            "Standard prefix with event type",
			prefix:          "user.service.",
			eventType:       "access.check",
			expectedRouting: "user.service.access.check",
		},
		{
			name:            "Empty prefix",
			prefix:          "",
			eventType:       "role.create",
			expectedRouting: "role.create",
		},
		{
			name:            "Empty event type",
			prefix:          "events",
			eventType:       "",
			expectedRouting: "events",
		},
		{
			name:            "Empty both",
			prefix:          "",
			eventType:       "",
			expectedRouting: "",
		},
		{
			name:            "Single word prefix",
			prefix:          "access.",
			eventType:       "check",
			expectedRouting: "access.check",
		},
		{
			name:            "Multi-level event type",
			prefix:          "role.",
			eventType:       "permission.assign",
			expectedRouting: "role.permission.assign",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			publisher := &RabbitmqPublisher{
				routingKeyPrefix: tt.prefix,
			}

			routingKey := publisher.getRoutingKey(tt.eventType)

			assert.Equal(t, tt.expectedRouting, routingKey)
		})
	}
}
