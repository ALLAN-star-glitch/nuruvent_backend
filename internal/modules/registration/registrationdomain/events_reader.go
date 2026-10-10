package registrationdomain

import "context"

// EventsReader resolves a team to its event IDs so the registration
// list can be team-scoped.
type EventsReader interface {
	ListEventIDsByTeam(ctx context.Context, userID, teamID string) ([]string, error)
}