// internal/modules/events/domain/team_repository.go (new file)

package domain

import "context"

// TeamLookup is a minimal read-only interface for resolving teams to
// accounts. The events module depends on this narrow contract instead of
// importing the full team module.
type TeamLookup interface {
	AccountIDForTeam(ctx context.Context, teamID string) (string, error)
}