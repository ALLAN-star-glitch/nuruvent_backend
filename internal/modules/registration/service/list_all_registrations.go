package service

import (
	"context"
	"fmt"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/registrationdomain"
)

// ListAllRegistrations returns registrations across every event
// owned by any account the caller belongs to.
//
// Ownership resolution:
//   1. Ask the permission checker for the caller's account IDs.
//   2. The repository filters events through
//      event_registrations.event_id → events.team_id → teams.account_id.
//
// If the caller has no accounts, returns an empty page — no error.
func (s *service) ListAllRegistrations(
	ctx context.Context,
	cmd ListAllRegistrationsCommand,
) (*ListAllRegistrationsResult, error) {

	if cmd.UserID == "" {
		return nil, registrationdomain.ErrNotOwner
	}
	if s.deps.PermissionChecker == nil {
		return nil, fmt.Errorf("permission checker is not configured")
	}

	accountIDs, err := s.deps.PermissionChecker.GetUserAccountIDs(ctx, cmd.UserID)
	if err != nil {
		return nil, fmt.Errorf("resolve accounts: %w", err)
	}

	// Normalize paging before the empty-accounts short-circuit so the
	// returned shape is consistent either way.
	page := cmd.Page
	if page < 1 {
		page = 1
	}
	pageSize := cmd.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// Empty accounts → empty page. Avoids an unscoped repository query.
	if len(accountIDs) == 0 {
		return &ListAllRegistrationsResult{
			Registrations: []*registrationdomain.CrossEventRegistrationRow{},
			Total:         0,
			Page:          page,
			PageSize:      pageSize,
		}, nil
	}

	// Parse statuses from raw slugs → domain Status values.
	statuses := make([]registrationdomain.Status, 0, len(cmd.Statuses))
	for _, raw := range cmd.Statuses {
		if parsed, err := registrationdomain.ParseStatus(raw); err == nil {
			statuses = append(statuses, parsed)
		}
	}

	filter := registrationdomain.ListAllFilter{
		AccountIDs: accountIDs,
		EventID:    cmd.EventID,
		Search:     cmd.Search,
		Statuses:   statuses,
		SortBy:     cmd.SortBy,
		SortOrder:  cmd.SortOrder,
		Page:       page,
		PageSize:   pageSize,
	}

	if err := s.applyTeamFilter(ctx, &filter, cmd.ActorID, cmd.TeamID, cmd.Scope); err != nil {
		return nil, err
	}

	registrations, total, err := s.deps.EventRegistrations.ListAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("list all registrations: %w", err)
	}

	// Ensure non-nil slice for JSON.
	if registrations == nil {
		registrations = []*registrationdomain.CrossEventRegistrationRow{}
	}

	return &ListAllRegistrationsResult{
		Registrations: registrations,
		Total:         total,
		Page:          page,
		PageSize:      pageSize,
	}, nil
}

// applyTeamFilter expands a team ID into event IDs and, when
// non-empty, narrows the filter to those events.
//
// Behaviour:
//   - Scope != "team" or TeamID == ""   → no-op, filter unchanged.
//   - TeamID set but ActorID empty      → ErrTeamAccessDenied.
//   - Team has zero events              → filter matches nothing.
//   - Team not found or caller denied   → ErrTeamAccessDenied.
func (s *service) applyTeamFilter(
	ctx context.Context,
	f *registrationdomain.ListAllFilter,
	actorID, teamID, scope string,
) error {
	if scope != "team" || teamID == "" {
		return nil
	}
	if actorID == "" {
		return registrationdomain.ErrTeamAccessDenied
	}
	if s.deps.EventsReader == nil {
		return fmt.Errorf("events reader not configured")
	}

	eventIDs, err := s.deps.EventsReader.ListEventIDsByTeam(ctx, actorID, teamID)
	if err != nil {
		return err
	}

	if len(eventIDs) == 0 {
		f.EventIDs = []string{"00000000-0000-0000-0000-000000000000"}
		return nil
	}

	f.EventIDs = eventIDs
	return nil
}