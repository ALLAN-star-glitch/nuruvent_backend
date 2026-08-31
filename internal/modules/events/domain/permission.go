// internal/modules/events/domain/permission.go

package domain

import "context"

type PermissionChecker interface {
    // CanManageEvent checks if user can manage events for an institution
    CanManageEvent(ctx context.Context, userID, institutionID string) bool
    
    // CanManagePersonalEvent checks if user can manage their own personal events
    CanManagePersonalEvent(ctx context.Context, userID string) bool
    
    // CanUpdateEvent checks if user can update events for a team
    CanUpdateEvent(ctx context.Context, userID, teamID string) bool
    
    // CanDeleteEvent checks if user can delete events for a team
    CanDeleteEvent(ctx context.Context, userID, teamID string) bool
    
    // CanViewEvent checks if user can view events for a team
    CanViewEvent(ctx context.Context, userID, teamID string) bool
}