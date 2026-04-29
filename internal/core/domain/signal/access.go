package signal

type SignalAccessCheck struct {
	SubjectID   string
	SubjectType string
	Resource    string
	Action      string
	Allowed     *bool
	Reason      *string
}
