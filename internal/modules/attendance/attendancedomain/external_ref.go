package attendancedomain

import "strings"

// ExternalRef identifies the source module entity that owns an
// attendee, session, or roll-up. The attendance module never
// interprets these values — it only stores and returns them.
//
// Examples of Type values chosen by consumers:
//
//	"event_registration"  — an attendee for an event
//	"event"               — a session's parent entity (single- or multi-session event)
//	"course_enrolment"    — an attendee for a course
//	"course_cohort"       — a session's parent entity (course cohort)
//
// The attendance module does not branch on Type.
type ExternalRef struct {
	Type string
	ID   string
}

func (r ExternalRef) IsValid() bool {
	return strings.TrimSpace(r.Type) != "" && strings.TrimSpace(r.ID) != ""
}

func (r ExternalRef) Equals(other ExternalRef) bool {
	return r.Type == other.Type && r.ID == other.ID
}