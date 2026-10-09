package service

import (
	"context"
	"fmt"
	"log"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
)

// ListAttendees returns a paginated attendee directory across every
// event owned by any account the caller belongs to.
//
// Ownership resolution:
//   1. Ask the permission checker for the caller's account IDs.
//   2. If scope == "team" + team_id is set, resolve the team's
//      event IDs through the events port and intersect.
//   3. The repository filters events through
//      events.team_id → teams.account_id.
//
// If the caller has no accounts, returns an empty page — no error.
// If the team scope resolves to zero events, also returns an empty
// page without hitting the attendee query.
func (s *attendanceService) ListAttendees(
	ctx context.Context,
	cmd ListAttendeesCommand,
) (*ListAttendeesResult, error) {

	if cmd.UserID == "" {
		return nil, attendance.ErrInvalidSession
	}
	if s.deps.PermissionChecker == nil {
		return nil, fmt.Errorf("permission checker is not configured")
	}

	accountIDs, err := s.deps.PermissionChecker.GetUserAccountIDs(ctx, cmd.UserID)
	if err != nil {
		return nil, fmt.Errorf("resolve accounts: %w", err)
	}
	

	log.Printf("[ListAttendees] user=%s team=%s scope=%s accounts=%d eventsReader=%v",
    cmd.UserID, cmd.TeamID, cmd.Scope, len(accountIDs), s.deps.EventsReader != nil)

	// Empty accounts → empty page. Short-circuit before hitting the repo.
	if len(accountIDs) == 0 {
		return emptyAttendeesPage(cmd.Page, cmd.PageSize), nil
	}

	// Team scope: resolve team → event IDs.
	var eventIDs []string
	if cmd.Scope == "team" && cmd.TeamID != "" {
		if s.deps.EventsReader != nil {
			ids, err := s.deps.EventsReader.ListEventIDsByTeam(ctx, cmd.UserID, cmd.TeamID)
			log.Printf("[ListAttendees] port returned: ids=%d err=%v", len(ids), err)
			if err != nil {
				return nil, fmt.Errorf("resolve team events: %w", err)
			}
			if len(ids) == 0 {
				return emptyAttendeesPage(cmd.Page, cmd.PageSize), nil
			}
			eventIDs = ids
		}
		// If EventsReader is nil, we silently fall through to the
		// account-scoped query. Wire the adapter to enable team scope.
	}

	q := attendance.ListAttendeesQuery{
		AccountIDs: accountIDs,
		EventIDs:   eventIDs,
		EventID:    cmd.EventID,
		Search:     cmd.Search,
		Statuses:   cmd.Statuses,
		SortBy:     cmd.SortBy,
		SortOrder:  cmd.SortOrder,
		Page:       cmd.Page,
		PageSize:   cmd.PageSize,
	}

	var res *attendance.ListAttendeesResult
	err = s.deps.UnitOfWork.Do(ctx, func(repos attendance.Repositories) error {
		r, err := repos.RollupStatuses.ListAttendees(ctx, q)
		if err != nil {
			return fmt.Errorf("list attendees: %w", err)
		}
		res = r
		return nil
	})
	if err != nil {
		return nil, err
	}

	out := &ListAttendeesResult{
		Attendees: make([]*CrossEventAttendeeItem, 0, len(res.Attendees)),
		Total:     res.Total,
		Page:      res.Page,
		PageSize:  res.PageSize,
	}
	for _, a := range res.Attendees {
		out.Attendees = append(out.Attendees, &CrossEventAttendeeItem{
			AttendeeID:           a.AttendeeID,
			DisplayName:          a.DisplayName,
			Email:                a.Email,
			Phone:                a.Phone,
			EventID:              a.EventID,
			EventName:            a.EventName,
			EventSlug:            a.EventSlug,
			EventStartDate:       a.EventStartDate,
			EffectiveStatus:      attendance.AttendanceStatus(a.DerivedStatus),
			SessionsTotal:        a.SessionsTotal,
			SessionsAttended:     a.SessionsAttended,
			SessionsConfirmed:    a.SessionsConfirmed,
			TotalDurationSeconds: a.TotalDurationSeconds,
			RegisteredAt:         a.RegisteredAt,
			LastActivityAt:       a.LastDerivedAt,
			IsHost:               a.IsHost,
		})
	}
	return out, nil
}

func emptyAttendeesPage(page, pageSize int) *ListAttendeesResult {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	return &ListAttendeesResult{
		Attendees: []*CrossEventAttendeeItem{},
		Total:     0,
		Page:      page,
		PageSize:  pageSize,
	}
}