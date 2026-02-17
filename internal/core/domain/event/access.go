package event

type EventAccessCheckData struct {
	SubjectId   string
	SubjectType string
	Resource    string
	Action      string
	Reason      string
}
