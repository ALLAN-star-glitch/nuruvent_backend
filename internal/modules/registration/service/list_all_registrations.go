package service

import (
	"context"
	"fmt"
	"log"

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




		// Own-scope decision.
	//
	// If the request carries a team, the operation is about one
	// account — authorize against that account only.
	//
	// If there is no team, the query spans every account the caller
	// belongs to. Skip the filter only if they hold registration:read
	// everywhere.
	var canReadAll bool
	var teamAccountID string

	if cmd.TeamID != "" {
		teamAccountID, err = s.deps.EventRegistrations.ResolveTeamAccount(ctx, cmd.TeamID)
		if err != nil {
			return nil, fmt.Errorf("resolve team account: %w", err)
		}
		canReadAll = s.callerHasReadInAccount(ctx, cmd.UserID, teamAccountID)
	} else {
		canReadAll = s.callerHasReadInEveryAccount(ctx, cmd.UserID, accountIDs)
	}

		if !canReadAll {
		// Not account-wide read. Two possibilities:
		//
		//   - Trainer: holds event:read_all → their scope is every
		//     registration on the teams they belong to. They run
		//     events on those teams, so they need operational
		//     visibility into who is registering. Allowed.
		//   - Learner: holds neither → the team registrations view
		//     is not for them. Deny with 403.
		//
		// Distinguishing requires knowing which account the view is
		// about. Without a team scope, we cannot tell trainer from
		// learner, so we fall through to denial — a caller with no
		// account-wide read has no legitimate reason to hit the
		// cross-account organizer view.
		//
		// NOTE: no EventOwnerUserID filter is applied for trainers.
		// The account + team scope from the filter is the boundary.
		// A trainer on a team sees that team's registrations; a
		// trainer on multiple teams sees the union. Payments remain
		// own-scoped via payment:read_own — this relaxation is
		// specific to registration and attendee visibility.
		isTrainer := teamAccountID != "" &&
			s.callerHoldsEventReadAll(ctx, cmd.UserID, teamAccountID)

		if !isTrainer {
			return nil, registrationdomain.ErrPermissionDenied
		}
	}


		log.Printf("🔍 ListAllRegistrations user=%s accounts=%d scoped_user=%q team_id=%q scope=%q can_read_all=%v",
		cmd.UserID, len(accountIDs), filter.UserID, cmd.TeamID, cmd.Scope, canReadAll)

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



// callerHasReadInAccount reports whether the caller holds the
// account-wide registration:read grant in one specific account.
func (s *service) callerHasReadInAccount(
	ctx context.Context,
	userID, accountID string,
) bool {
	if s.deps.PermissionChecker == nil || accountID == "" {
		return false
	}
	domain := "account:" + accountID
	allowed, err := s.deps.PermissionChecker.HasPermission(
		ctx, userID, domain, "registration", "read",
	)
	return err == nil && allowed
}

// callerHasReadInEveryAccount reports whether the caller holds
// registration:read in every account they belong to.
//
// Used only when the request has no team scope — in that case the
// query spans all accounts, and the own-scope filter is skipped only
// if the caller is an admin everywhere.
func (s *service) callerHasReadInEveryAccount(
	ctx context.Context,
	userID string,
	accountIDs []string,
) bool {
	if len(accountIDs) == 0 {
		return false
	}
	for _, id := range accountIDs {
		if !s.callerHasReadInAccount(ctx, userID, id) {
			return false
		}
	}
	return true
}

// callerHoldsEventReadAll reports whether the caller holds the
// account-wide event:read_all grant in a specific account.
//
// Used to distinguish trainers from learners: trainers hold
// event:read_all (they see the whole account's event roster),
// learners hold event:read_own. Both hold registration:read_own, so
// the registration grant alone is not enough to tell them apart.
func (s *service) callerHoldsEventReadAll(
	ctx context.Context,
	userID, accountID string,
) bool {
	if s.deps.PermissionChecker == nil || accountID == "" {
		return false
	}
	domain := "account:" + accountID
	allowed, err := s.deps.PermissionChecker.HasPermission(
		ctx, userID, domain, "event", "read_all",
	)
	return err == nil && allowed
}