package attendancedomain

import (
	"fmt"
	"strings"
	"time"
)

// AttendanceOverride is an audit record of a manual status change by
// a host.
type AttendanceOverride struct {
	ID          string
	AttendeeID  string
	SessionID   string
	ActorID     string
	PriorStatus AttendanceStatus
	NewStatus   AttendanceStatus
	Reason      string
	CreatedAt   time.Time
}

// NewAttendanceOverride constructs a new override.
func NewAttendanceOverride(
	id, attendeeID, sessionID, actorID string,
	prior, next AttendanceStatus,
	reason string,
	now time.Time,
) (*AttendanceOverride, error) {
	if strings.TrimSpace(id) == "" ||
		strings.TrimSpace(attendeeID) == "" ||
		strings.TrimSpace(sessionID) == "" ||
		strings.TrimSpace(actorID) == "" {
		return nil, fmt.Errorf("%w: ids are required", ErrInvalidAttendance)
	}
	if prior == next {
		return nil, fmt.Errorf("%w: no-op override", ErrInvalidAttendance)
	}
	return &AttendanceOverride{
		ID:          id,
		AttendeeID:  attendeeID,
		SessionID:   sessionID,
		ActorID:     actorID,
		PriorStatus: prior,
		NewStatus:   next,
		Reason:      reason,
		CreatedAt:   now,
	}, nil
}

// HydrateAttendanceOverride reconstructs an override from
// persistence.
func HydrateAttendanceOverride(
	id, attendeeID, sessionID, actorID string,
	prior, next AttendanceStatus,
	reason string,
	createdAt time.Time,
) *AttendanceOverride {
	return &AttendanceOverride{
		ID:          id,
		AttendeeID:  attendeeID,
		SessionID:   sessionID,
		ActorID:     actorID,
		PriorStatus: prior,
		NewStatus:   next,
		Reason:      reason,
		CreatedAt:   createdAt,
	}
}