// internal/modules/events/domain/attendance_reader.go

package domain

import "context"

// ============================================================
// OUTBOUND PORT: AttendanceReader
// ============================================================
//
// Reads attendance data for an event. Implemented by a cross-module
// adapter that delegates to the attendance service.
//
// The events module uses this to determine who attended an event, so
// certificates can be issued to qualified recipients without the
// caller having to supply a manual list.
type AttendanceReader interface {
	// GetAttendedRecipients returns every attendee for the given event
	// whose derived attendance status qualifies them for a certificate.
	//
	// Hosts are excluded — they organized the event, they don't
	// complete it. Admins who want a host to receive a certificate
	// can add them manually via the "add" mode.
	//
	// Qualification is defined by the attendance module. The events
	// module treats the returned list as authoritative.
	GetAttendedRecipients(
		ctx context.Context,
		eventID string,
	) ([]CertificateRecipient, error)
}