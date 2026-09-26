package attendancedomain

import (
	"fmt"
	"strings"
	"time"
)

// AttendanceRecord is one join/leave interval for an attendee in a
// session. Multiple records per (attendee, session) are allowed and
// expected — they represent reconnects, network drops, or dual device
// usage.
type AttendanceRecord struct {
	ID              string
	AttendeeID      string
	SessionID       string
	JoinTime        time.Time
	LeaveTime       *time.Time
	DurationSeconds int64
	Source          AttendanceSource
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// NewAttendanceRecord constructs a new attendance record.
func NewAttendanceRecord(
	id, attendeeID, sessionID string,
	joinTime time.Time,
	source AttendanceSource,
	now time.Time,
) (*AttendanceRecord, error) {
	if strings.TrimSpace(id) == "" || strings.TrimSpace(attendeeID) == "" || strings.TrimSpace(sessionID) == "" {
		return nil, fmt.Errorf("%w: ids are required", ErrInvalidAttendance)
	}
	if !source.IsValid() {
		return nil, fmt.Errorf("%w: invalid source %q", ErrInvalidAttendance, source)
	}
	return &AttendanceRecord{
		ID:         id,
		AttendeeID: attendeeID,
		SessionID:  sessionID,
		JoinTime:   joinTime,
		Source:     source,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

// HydrateAttendanceRecord reconstructs a record from persistence.
func HydrateAttendanceRecord(
	id, attendeeID, sessionID string,
	joinTime time.Time,
	leaveTime *time.Time,
	durationSeconds int64,
	source AttendanceSource,
	createdAt, updatedAt time.Time,
) *AttendanceRecord {
	return &AttendanceRecord{
		ID:              id,
		AttendeeID:      attendeeID,
		SessionID:       sessionID,
		JoinTime:        joinTime,
		LeaveTime:       leaveTime,
		DurationSeconds: durationSeconds,
		Source:          source,
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
	}
}

// IsOpen reports whether the record is still active.
func (r *AttendanceRecord) IsOpen() bool {
	return r.LeaveTime == nil
}

// Close marks the record as ended, computing the duration. Idempotent.
func (r *AttendanceRecord) Close(leaveTime time.Time, now time.Time) error {
	if r.LeaveTime != nil {
		return nil
	}
	if !leaveTime.After(r.JoinTime) {
		return fmt.Errorf("%w: leave must be after join", ErrInvalidAttendance)
	}
	r.LeaveTime = &leaveTime
	r.DurationSeconds = int64(leaveTime.Sub(r.JoinTime).Seconds())
	r.UpdatedAt = now
	return nil
}