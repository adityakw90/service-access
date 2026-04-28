package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubjectProfile(t *testing.T) {
	profile := &SubjectProfile{
		Groups:      []Group{},
		Roles:       []Role{},
		Permissions: []Permission{},
	}
	assert.Equal(t, 0, len(profile.Groups))
	assert.Equal(t, 0, len(profile.Roles))
	assert.Equal(t, 0, len(profile.Permissions))
}
