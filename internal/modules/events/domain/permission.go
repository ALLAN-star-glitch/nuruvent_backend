// internal/modules/events/domain/permission.go

package domain

import "context"

// PermissionChecker defines what the events domain needs for authorization
// The teamID can be either a user_id (personal team) or institution_id (institution team)
type PermissionChecker interface {
	// CanManageEvent checks if user can manage events for a team
	CanManageEvent(ctx context.Context, userID, teamID string) bool

	// CanUpdateEvent checks if user can update events for a team
	CanUpdateEvent(ctx context.Context, userID, teamID string) bool

	// CanDeleteEvent checks if user can delete events for a team
	CanDeleteEvent(ctx context.Context, userID, teamID string) bool

	// CanViewEvent checks if user can view events for a team
	CanViewEvent(ctx context.Context, userID, teamID string) bool
}