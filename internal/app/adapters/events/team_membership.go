// internal/app/adapters/events/team_membership.go
package events

import (
	"context"

	eventsDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
	teamService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/team/service"
)

type teamMembershipAdapter struct {
	teams teamService.Service
}

// NewTeamMembershipAdapter bridges the team service to the events
// module's TeamMembershipChecker port.
func NewTeamMembershipAdapter(
	teams teamService.Service,
) eventsDomain.TeamMembershipChecker {
	return &teamMembershipAdapter{teams: teams}
}

func (a *teamMembershipAdapter) IsTeamMember(
	ctx context.Context,
	teamID, userID string,
) (bool, error) {
	return a.teams.IsTeamMember(ctx, teamID, userID)
}

var _ eventsDomain.TeamMembershipChecker = (*teamMembershipAdapter)(nil)