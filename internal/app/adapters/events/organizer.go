package events

import (
	"context"
	"errors"
	"fmt"
	"strings"

	accountsvc "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/service"
	eventsdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
)

func normalizeAccountType(raw string) string {
    return strings.TrimPrefix(raw, "account-type-")
}

// OrganizerAdapter implements eventsdomain.OrganizerProvider by delegating
// to the account module's service.
//
// Lives in app/adapters/events/ because it exists to satisfy the events
// module's outbound port. The events module itself has no compile-time
// dependency on the account module — this adapter is the only glue.
type OrganizerAdapter struct {
	accounts accountsvc.Service
}

// NewOrganizerAdapter constructs the adapter.
// Returns the port interface type so callers depend on the abstraction.
func NewOrganizerAdapter(accounts accountsvc.Service) eventsdomain.OrganizerProvider {
	return &OrganizerAdapter{accounts: accounts}
}

func (a *OrganizerAdapter) GetOrganizer(ctx context.Context, teamID string) (*eventsdomain.OrganizerInfo, error) {
	if teamID == "" {
		return nil, errors.New("team ID is required")
	}

	account, err := a.accounts.GetAccountByTeamID(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve account for team %s: %w", teamID, err)
	}
	if account == nil {
		return nil, nil // no organizer
	}

	return &eventsdomain.OrganizerInfo{
		ID:          account.ID,
		Type:        normalizeAccountType(account.Type),
		Name:        account.Name,
		DisplayName: account.DisplayName,
		Slug:        account.Slug,
		AvatarURL:   account.LogoURL,
	}, nil
}