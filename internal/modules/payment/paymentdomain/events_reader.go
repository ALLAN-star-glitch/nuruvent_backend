package paymentdomain

import "context"

// ============================================================
// EVENTS READER PORT
// ============================================================
//
// The payment module uses this port to resolve a team to its event
// IDs, so the ledger and stats can be filtered by team.
//
// The attendance module has an identical port
// (attendancedomain.EventsReader). They're intentionally separate —
// each module owns its own port surface and neither imports the
// other's domain.
//
// The adapter in internal/app/adapters/payment/events_reader.go
// implements this on top of the events service.

type EventsReader interface {
	// ListEventIDsByTeam returns the IDs of every event belonging to
	// teamID, subject to the caller's permissions. Returns
	// ErrTeamAccessDenied if the caller isn't authorized to see the
	// team's events.
	//
	// The userID is passed explicitly (not via context) because
	// context propagation across module boundaries is fragile — the
	// receiving module doesn't know what keys the sender put in.
	ListEventIDsByTeam(ctx context.Context, userID, teamID string) ([]string, error)
}