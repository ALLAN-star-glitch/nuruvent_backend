package events

import (
	"context"
	"fmt"

	attendanceDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
	attendanceService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/service"
	eventsDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
)

// AttendanceRegistrarAdapter implements the events module's
// AttendanceRegistrar port by delegating to the attendance service.
type AttendanceRegistrarAdapter struct {
	attendance attendanceService.Service
}

// NewAttendanceRegistrarAdapter wires the events module's
// AttendanceRegistrar port to the attendance service.
func NewAttendanceRegistrarAdapter(
	attendance attendanceService.Service,
) eventsDomain.AttendanceRegistrar {
	return &AttendanceRegistrarAdapter{attendance: attendance}
}

// UpsertSession mirrors one event schedule into the attendance module.
func (a *AttendanceRegistrarAdapter) UpsertSession(
	ctx context.Context,
	cmd eventsDomain.AttendanceUpsertSessionCommand,
) error {
	provider := attendanceDomain.SessionProvider(cmd.Provider)
	if !provider.IsValid() {
		return fmt.Errorf("invalid provider %q for schedule %s",
			cmd.Provider, cmd.ProviderSessionID)
	}

	_, err := a.attendance.UpsertSession(ctx, attendanceService.UpsertSessionCommand{
		External: attendanceDomain.ExternalRef{
			Type: cmd.ExternalType,
			ID:   cmd.ExternalID,
		},
		ProviderSessionID: cmd.ProviderSessionID,
		Title:             cmd.Title,
		ScheduledStart:    cmd.ScheduledStart,
		ScheduledEnd:      cmd.ScheduledEnd,
		Provider:          provider,
		ProviderMeetingID: cmd.ProviderMeetingID,
		ProviderURL:       cmd.ProviderURL,
	})
	if err != nil {
		return fmt.Errorf("attendance.UpsertSession: %w", err)
	}
	return nil
}