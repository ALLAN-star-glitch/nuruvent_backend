package domain

import (
	"context"
	"time"
)

// AttendanceRegistrar is the events module's outbound port for
// attendance sync. Implemented by a cross-module adapter that
// delegates to the attendance service.
//
// See internal/app/adapters/events/attendance_registrar.go.
type AttendanceRegistrar interface {
	// UpsertSession creates or updates an attendance session for a
	// single event schedule.
	//
	// Idempotent on (ExternalType, ExternalID, ProviderSessionID).
	UpsertSession(ctx context.Context, cmd AttendanceUpsertSessionCommand) error
}

// AttendanceUpsertSessionCommand carries the minimum data the
// attendance module needs to track one session.
type AttendanceUpsertSessionCommand struct {
	// ExternalType identifies the kind of parent entity.
	// Always "event" from this module.
	ExternalType string

	// ExternalID is the parent event's UUID.
	ExternalID string

	// ProviderSessionID is the event_schedules row's UUID. This is
	// what disambiguates one session from another within the parent.
	ProviderSessionID string

	// Title is a human-readable label for the session.
	Title string

	// ScheduledStart and ScheduledEnd are absolute times in UTC.
	ScheduledStart time.Time
	ScheduledEnd   time.Time

	// Provider is one of: "zoom", "google_meet", "in_person", "none".
	Provider string

	// ProviderMeetingID is the extracted meeting ID. Empty for
	// in-person and none.
	ProviderMeetingID string

	// ProviderURL is the join URL (Zoom link, Meet link, or
	// in-person location string).
	ProviderURL string
}