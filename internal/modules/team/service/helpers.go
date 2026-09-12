// internal/modules/team/service/helpers.go

package service

import (
	"context"
	"fmt"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/accountdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/team/teamdomain"
)

// ============================================================
// CONTEXT KEYS
// ============================================================

// contextKey is an unexported type for context keys in this package, to
// avoid collisions with context keys defined in other packages.
type contextKey string

// UserContextKey is the key under which the authenticated user's ID is
// stored in the request context.
//
// The delivery layer (HTTP handler or middleware) is responsible for
// setting it before calling service methods that authorize against the
// actor. See withActor for the setter.
const UserContextKey contextKey = "user_id"

// ============================================================
// CONTEXT HELPERS
// ============================================================

// actorFromContext returns the authenticated user ID from ctx, or "" if
// it's not set.
//
// Service methods that need to authorize an action call this and treat
// an empty result as ErrPermissionDenied.
func actorFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(UserContextKey).(string); ok {
		return v
	}
	return ""
}

// withActor returns a copy of ctx with the given user ID stored under
// UserContextKey. Useful in tests and in the delivery layer.
func WithActor(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserContextKey, userID)
}

// ============================================================
// ACCOUNT DOMAIN HELPERS
// ============================================================

// accountDomainForTeam returns the Casbin account domain for a team.
//
// POST-REVAMP: teams are not authorization domains. Every team-scoped
// permission check is performed against the team's parent ACCOUNT domain:
//
//	account:<team.account_id>
//
// The team must have a non-empty AccountID (the database enforces this via
// a NOT NULL constraint on teams.account_id).
//
// Returns an error if the team is nil or has no valid account.
func (s *teamService) accountDomainForTeam(team *teamdomain.Team) (string, error) {
	if team == nil {
		return "", fmt.Errorf("team is nil")
	}
	if team.AccountID == "" {
		return "", fmt.Errorf("team %s has no account ID", team.ID)
	}
	domain := accountdomain.AccountDomain(team.AccountID)
	if domain == "" {
		return "", fmt.Errorf("team %s produced an empty account domain", team.ID)
	}
	return domain, nil
}

// accountDomainForTeamID loads the team and returns its account domain.
//
// Convenience for callers that only have a team ID. Performs a DB read;
// prefer accountDomainForTeam when you already have the team loaded.
func (s *teamService) accountDomainForTeamID(ctx context.Context, teamID string) (string, error) {
	team, err := s.repo.GetTeamByID(ctx, teamID)
	if err != nil {
		return "", err
	}
	if team == nil {
		return "", teamdomain.ErrTeamNotFound
	}
	return s.accountDomainForTeam(team)
}