package service

import (
	"context"
)

type AccessService interface {
	// Check checks if a subject has access to a resource
	Check(ctx context.Context, subjectID string, subjectType string, resource string, action string) (allowed bool, reason string, err error)
}
