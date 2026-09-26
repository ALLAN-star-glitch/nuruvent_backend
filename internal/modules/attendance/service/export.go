// internal/modules/attendance/service/export.go

package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strconv"
	"time"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
)

// ExportSessionAttendance returns a CSV for one session's attendance.
//
// The CSV columns are:
//
//	name,email,status,duration_seconds,host_confirmed,last_updated
//
// `status` is the effective status (host override wins if present).
// `duration_seconds` is the total seconds the attendee was present.
// `last_updated` is the RFC 3339 timestamp of the last derivation.
//
// The output is suitable for sharing with:
//   - Institutional reporting teams.
//   - Professional bodies verifying CPD attendance.
//   - Event organizers who want a spreadsheet.
//
// Uses RFC 4180 encoding: fields with commas or quotes are properly
// escaped.
func (s *attendanceService) ExportSessionAttendance(
	ctx context.Context,
	sessionID string,
) ([]byte, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("session_id is required")
	}

	type row struct {
		name           string
		email          string
		status         attendance.AttendanceStatus
		durationSec    int64
		hostConfirmed  bool
		lastDerivedAt  time.Time
	}

	var rows []row

	txErr := s.deps.UnitOfWork.Do(ctx, func(repos attendance.Repositories) error {
		// Verify session exists.
		if _, err := repos.Sessions.FindByID(ctx, sessionID); err != nil {
			return fmt.Errorf("load session: %w", err)
		}

		statuses, err := repos.SessionStatuses.ListBySession(ctx, sessionID)
		if err != nil {
			return fmt.Errorf("list statuses: %w", err)
		}

		for _, st := range statuses {
			attendee, err := repos.Attendees.FindByID(ctx, st.AttendeeID)
			if err != nil {
				return fmt.Errorf("load attendee %s: %w", st.AttendeeID, err)
			}
			rows = append(rows, row{
				name:          attendee.DisplayName,
				email:         attendee.Email,
				status:        st.EffectiveStatus(),
				durationSec:   st.TotalDurationSeconds,
				hostConfirmed: st.HostConfirmed,
				lastDerivedAt: st.LastDerivedAt,
			})
		}
		return nil
	})
	if txErr != nil {
		return nil, txErr
	}

	// Build the CSV.
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	// Header.
	if err := w.Write([]string{
		"name",
		"email",
		"status",
		"duration_seconds",
		"host_confirmed",
		"last_updated",
	}); err != nil {
		return nil, fmt.Errorf("write csv header: %w", err)
	}

	// Rows.
	for _, r := range rows {
		if err := w.Write([]string{
			r.name,
			r.email,
			string(r.status),
			strconv.FormatInt(r.durationSec, 10),
			strconv.FormatBool(r.hostConfirmed),
			r.lastDerivedAt.UTC().Format(time.RFC3339),
		}); err != nil {
			return nil, fmt.Errorf("write csv row: %w", err)
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, fmt.Errorf("flush csv: %w", err)
	}

	return buf.Bytes(), nil
}