package event

type EventCheckAccessData struct {
	SubjectId   string
	SubjectType string
	Resource    string
	Action      string
	Reason      string
}

type EventCheckAccessFailedData struct {
	SubjectId   string
	SubjectType string
	Resource    string
	Action      string
	Reason      string
}
