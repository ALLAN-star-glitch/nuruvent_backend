// internal/app/adapters/registration/attendance_registrar.go

package registration

import (
	"context"
	"fmt"
	"time"

	attendanceDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
	attendanceService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/service"
	registrationDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/registrationdomain"
)

// AttendanceRegistrarAdapter implements the registration module's
// AttendanceRegistrar port by delegating to the attendance service.
type AttendanceRegistrarAdapter struct {
	attendance    attendanceService.Service
	publicBaseURL string
}

func NewAttendanceRegistrarAdapter(
	attendance attendanceService.Service,
	publicBaseURL string,
) registrationDomain.AttendanceRegistrar {
	return &AttendanceRegistrarAdapter{
		attendance:    attendance,
		publicBaseURL: publicBaseURL,
	}
}

func (a *AttendanceRegistrarAdapter) RegisterAttendee(
	ctx context.Context,
	cmd registrationDomain.RegisterAttendeeForAttendanceCommand,
) (string, error) {
	attendee, err := a.attendance.RegisterAttendee(ctx, attendanceService.RegisterAttendeeCommand{
		External: attendanceDomain.ExternalRef{
			Type: cmd.ExternalType,
			ID:   cmd.ExternalID,
		},
		DisplayName: cmd.DisplayName,
		Email:       cmd.Email,
	})
	if err != nil {
		return "", fmt.Errorf("attendance.RegisterAttendee: %w", err)
	}
	return attendee.ID, nil
}

// RegisterAttendeeForEvent creates status rows for every session
// under the event and issues a join token per session.
//
// The returned join URLs are discarded here — this port method is
// used by the registration flow to ensure the attendee is on every
// roster. Callers that need the URLs use IssueJoinTokens instead.
func (a *AttendanceRegistrarAdapter) RegisterAttendeeForEvent(
	ctx context.Context,
	cmd registrationDomain.RegisterAttendeeForEventCommand,
) error {
	_, err := a.attendance.RegisterAttendeeForExternal(
		ctx,
		attendanceService.RegisterAttendeeForExternalCommand{
			AttendeeID: cmd.AttendeeID,
			External: attendanceDomain.ExternalRef{
				Type: "event",
				ID:   cmd.EventID,
			},
			PublicBaseURL: a.publicBaseURL,
			LinkGrace:     24 * time.Hour,
		},
	)
	if err != nil {
		return fmt.Errorf("attendance.RegisterAttendeeForExternal: %w", err)
	}
	return nil
}

// IssueJoinTokens issues a join token per session under an event and
// returns the resulting links.
//
// The URL comes from the attendance service's own result rather than
// string concatenation, so the attendance module owns the URL shape
// and the caller just consumes it.
func (a *AttendanceRegistrarAdapter) IssueJoinTokens(
	ctx context.Context,
	cmd registrationDomain.IssueJoinTokensCommand,
) ([]registrationDomain.JoinLink, error) {
	sessions, err := a.attendance.ListSessionsForExternal(ctx, attendanceDomain.ExternalRef{
		Type: "event",
		ID:   cmd.EventID,
	})
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}

	links := make([]registrationDomain.JoinLink, 0, len(sessions))
	for _, session := range sessions {
		// In-person sessions have no remote meeting to join. Skip
		// them entirely — no token, no link, no error.
		if !session.Provider.RequiresMeetingID() {
			continue
		}

		result, err := a.attendance.IssueJoinToken(ctx, attendanceService.IssueJoinTokenCommand{
			AttendeeID:    cmd.AttendeeID,
			SessionID:     session.ID,
			Grace:         24 * time.Hour,
			PublicBaseURL: a.publicBaseURL,
		})
		if err != nil {
			return nil, fmt.Errorf("issue join token for session %s: %w", session.ID, err)
		}
		links = append(links, registrationDomain.JoinLink{
			SessionID:      session.ID,
			SessionTitle:   session.Title,
			ScheduledStart: session.ScheduledStart,
			ScheduledEnd:   session.ScheduledEnd,
			Platform:       string(session.Provider),
			URL:            result.JoinURL,
			ExpiresAt:      result.ExpiresAt,
		})
	}
	return links, nil
}





