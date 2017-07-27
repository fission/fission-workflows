package eventids

import (
	"fmt"
	"github.com/fission/fission-workflow/pkg/eventstore"
	"strings"
)

const (
	SUBJECT_SEPARATOR = "."
	ID_SEPARATOR      = "#"
)

/*
Format of a stringified EventID: <subject>.<subject>#<id>
*/
func ParseString(stringifiedEventId string) (*eventstore.EventID, error) {
	eventId := &eventstore.EventID{}
	p := strings.Split(stringifiedEventId, ID_SEPARATOR)

	switch len(p) {
	case 1:
		eventId.Id = p[0]
	case 2:
		eventId.Id = p[1]
		eventId.Subjects = strings.Split(p[0], SUBJECT_SEPARATOR)
	default:
		return nil, fmt.Errorf("Could not parse invalid EventID '%s'", stringifiedEventId)
	}

	return eventId, nil
}

func ToString(eventId *eventstore.EventID) string {
	return fmt.Sprintf("%s%s%s",
		strings.Join(eventId.GetSubjects(), SUBJECT_SEPARATOR),
		ID_SEPARATOR,
		eventId.GetId())
}

func New(subjects []string, id string) *eventstore.EventID {
	return &eventstore.EventID{
		Subjects: subjects,
		Id:       id,
	}
}

func NewSubject(subjects ...string) *eventstore.EventID {
	return New(subjects, "")
}
