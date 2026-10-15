// internal/app/adapters/events/attendance_reader.go

package events

import (
	"context"

	attendanceService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/service"
	eventsDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
)

// certificateEligibleStatuses lists the attendance rollup statuses
// that qualify an attendee for a certificate.
//
//   - full      — attended every session
//   - confirmed — attendance explicitly confirmed
//   - partial   — attended at least one session of a multi-session event
//   - joined    — connected to a session (may still be in progress)
//
// "no-show" and "registered" mean the person never connected at all.
// Those do not qualify.
//
// v1 policy: include "joined" so organizers can issue certificates
// during or right after an event, before session-end rollups settle.
// Organizers can still uncheck anyone in the batch dialog.
//
// TODO: replace with a per-event policy when CPD-grade events need a
// stricter rule.
var certificateEligibleStatuses = []string{"full", "confirmed", "partial", "joined"}

// AttendanceReaderAdapter bridges the attendance service to the
// events module's AttendanceReader port.
type AttendanceReaderAdapter struct {
	svc attendanceService.Service
}

// NewAttendanceReaderAdapter constructs an AttendanceReaderAdapter.
func NewAttendanceReaderAdapter(svc attendanceService.Service) *AttendanceReaderAdapter {
	return &AttendanceReaderAdapter{svc: svc}
}

// GetAttendedRecipients implements eventsDomain.AttendanceReader.
func (a *AttendanceReaderAdapter) GetAttendedRecipients(
	ctx context.Context,
	eventID string,
) ([]eventsDomain.CertificateRecipient, error) {
	// Query the entire event roster. Certificate issuance operates on
	// the whole attendee list at once, so we use a large page size.
	const pageSize = 500

	result, err := a.svc.ListEventAttendees(ctx, attendanceService.ListEventAttendeesCommand{
		EventID:  eventID,
		Statuses: certificateEligibleStatuses,
		Page:     1,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, err
	}

	recipients := make([]eventsDomain.CertificateRecipient, 0, len(result.Attendees))
	for _, item := range result.Attendees {
		// Skip hosts — they're organizers, not participants.
		if item.IsHost {
			continue
		}
		recipients = append(recipients, eventsDomain.CertificateRecipient{
			Name:  item.DisplayName,
			Email: item.Email,
			// CompletedAt is left zero; the events service fills it
			// with the event's start date before passing to the
			// certificate module.
		})
	}
	return recipients, nil
}

// compile-time assertion
var _ eventsDomain.AttendanceReader = (*AttendanceReaderAdapter)(nil)