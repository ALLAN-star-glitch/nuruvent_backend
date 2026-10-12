package attendancedomain

import (
	"context"
)



// EventsReader is the driven port for reading event data owned by
// the events module.
type EventsReader interface {
	// ListEventIDsByTeam returns every event ID that belongs to a
	// team, scoped to the calling user. userID is required.
	ListEventIDsByTeam(ctx context.Context, userID, teamID string) ([]string, error)

	// ResolveEventAccount returns the account UUID that owns an event.
	//
	// Chain inside the events module: events.team_id → teams.account_id.
	// Returns an error when the event does not exist or has no team.
	ResolveEventAccount(ctx context.Context, eventID string) (string, error)
}