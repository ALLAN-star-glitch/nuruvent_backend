package attendancedomain

import "context"

// EventsReader is the driven port for reading event data owned by
// the events module.
//
// The attendance module uses this to resolve team → event IDs
// without joining against the events schema directly.
type EventsReader interface {
	// ListEventIDsByTeam returns every event ID that belongs to a
	// team, scoped to the calling user. userID is required — the
	// events service refuses to enumerate without it.
	ListEventIDsByTeam(ctx context.Context, userID, teamID string) ([]string, error)
}