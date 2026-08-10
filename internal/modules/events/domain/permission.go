package domain

import "context"

// PermissionChecker defines what the events domain needs for authorization
type PermissionChecker interface {
    CanManageEvent(ctx context.Context, userID, eventAccountID string) bool
    CanReadEvent(ctx context.Context, userID, eventAccountID string) bool
    CanUpdateEvent(ctx context.Context, userID, eventAccountID string) bool
    CanDeleteEvent(ctx context.Context, userID, eventAccountID string) bool
}