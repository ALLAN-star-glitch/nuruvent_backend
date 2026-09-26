// internal/modules/attendance/infrastructure/publisher/log_publisher.go

package publisher

import (
	"context"
	"log"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/service"
)

// LogPublisher is a StatusPublisher that logs status changes.
//
// Used until the notification and registration modules subscribe
// properly. Replace with an outbox-backed publisher when downstream
// consumers depend on the events.
type LogPublisher struct{}

// NewLogPublisher constructs a log-only publisher.
func NewLogPublisher() *LogPublisher {
	return &LogPublisher{}
}

// SessionStatusChanged logs a change to a single (attendee, session)
// status.
func (p *LogPublisher) SessionStatusChanged(
	ctx context.Context,
	c service.SessionStatusChange,
) error {
	log.Printf(
		"[attendance] session status changed: attendee=%s session=%s %s->%s",
		c.AttendeeID, c.SessionID, c.OldStatus, c.NewStatus,
	)
	return nil
}

// RollupStatusChanged logs a change to an attendee's roll-up status.
func (p *LogPublisher) RollupStatusChanged(
	ctx context.Context,
	c service.RollupStatusChange,
) error {
	log.Printf(
		"[attendance] rollup status changed: attendee=%s external=%s/%s %s->%s",
		c.AttendeeID, c.External.Type, c.External.ID, c.OldStatus, c.NewStatus,
	)
	return nil
}

// Compile-time assertion.
var _ service.StatusPublisher = (*LogPublisher)(nil)