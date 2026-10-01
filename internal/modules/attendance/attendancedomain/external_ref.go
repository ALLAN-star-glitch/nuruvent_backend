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


// EmailRequired reports whether attendees referenced by this type of
// external ref must have an email address.
//
// Registration-based sources (event registrations, course
// enrolments) always carry an email — the platform collected it at
// signup time. Platform observation sources (e.g. participants
// observed from a video provider's API) do not: the source platform
// may not expose an email at all.
//
// New attendee-referencing types should be added to the switch when
// the platform guarantees an email. Everything else defaults to
// "email not required", which is the permissive choice — the
// alternative would silently reject new sources.
func (r ExternalRef) EmailRequired() bool {
	switch r.Type {
	case "event_registration", "course_enrolment":
		return true
	default:
		return false
	}
}