package service

import "github.com/adityakw90/service-access/internal/core/port/security"

// MockUIDGenerator is a mock implementation of security.UIDGenerator for testing.
type MockUIDGenerator struct {
	MockNew func() string
}

// New generates a mock UID. If MockNew is nil, returns a default value.
func (m *MockUIDGenerator) New() string {
	if m.MockNew != nil {
		return m.MockNew()
	}
	return "mock-uid-123"
}

// Ensure MockUIDGenerator implements the interface at compile time.
var _ security.UIDGenerator = (*MockUIDGenerator)(nil)
