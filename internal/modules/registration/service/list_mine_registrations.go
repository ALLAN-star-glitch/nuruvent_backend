package service

import (
	"context"
	"fmt"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/registrationdomain"
)

// ListMineRegistrations returns the caller's own registrations.
//
// Unlike ListAllRegistrations, no account-ownership check is needed:
// the filter is scoped to the caller's own user_id.
func (s *service) ListMineRegistrations(
	ctx context.Context,
	cmd ListMineRegistrationsCommand,
) (*ListMineRegistrationsResult, error) {

	if cmd.UserID == "" {
		return nil, registrationdomain.ErrNotOwner
	}

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

	statuses := make([]registrationdomain.Status, 0, len(cmd.Statuses))
	for _, raw := range cmd.Statuses {
		if parsed, err := registrationdomain.ParseStatus(raw); err == nil {
			statuses = append(statuses, parsed)
		}
	}

	filter := registrationdomain.ListMineFilter{
		UserID:    cmd.UserID,
		EventID:   cmd.EventID,
		Search:    cmd.Search,
		Statuses:  statuses,
		SortBy:    cmd.SortBy,
		SortOrder: cmd.SortOrder,
		Page:      page,
		PageSize:  pageSize,
	}

	registrations, total, err := s.deps.EventRegistrations.ListMine(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("list my registrations: %w", err)
	}

	if registrations == nil {
		registrations = []*registrationdomain.CrossEventRegistrationRow{}
	}

	return &ListMineRegistrationsResult{
		Registrations: registrations,
		Total:         total,
		Page:          page,
		PageSize:      pageSize,
	}, nil
}