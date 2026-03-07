package model

// SubjectProfile aggregates all access information for a subject.
type SubjectProfile struct {
    Groups      []Group
    Roles       []Role
    Permissions []Permission
}
