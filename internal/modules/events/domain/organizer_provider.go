package domain

import "context"

// ============================================================
// OUTBOUND PORT: OrganizerProvider
// ============================================================
//
// Resolves the public-facing "organizer" of an event.
//
// In this system, every team belongs to an account. The accounts table is
// the unified identity table — it holds both personal accounts (users) and
// institution accounts. Therefore, resolving an organizer is always:
//
//     team_id  →  teams.account_id  →  accounts row
//
// The account's `type` field ("personal" | "institution") determines how
// the organizer should be presented (e.g. institution name vs. user name).
//
// The events module does not know or care about the underlying table layout;
// it only asks this port for the organizer info.

type OrganizerProvider interface {
	// GetOrganizer returns the public organizer info for the given team.
	// Returns (nil, nil) if the team does not exist (caller treats as no organizer).
	GetOrganizer(ctx context.Context, teamID string) (*OrganizerInfo, error)
}