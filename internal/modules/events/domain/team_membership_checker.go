// internal/modules/events/domain/team_membership_checker.go
package domain

import "context"

// TeamMembershipChecker is the driven port for verifying whether a user
// belongs to a team.
//
// Used by the events service to gate access to private events: only
// members of the team that owns the event may view it.
//
// The adapter lives in internal/app/adapters/events/ and delegates to
// the team service. The events module never touches team_members
// directly.
type TeamMembershipChecker interface {
	IsTeamMember(ctx context.Context, teamID, userID string) (bool, error)
}