package domain

import "context"

// ============================================================
// OUTBOUND PORT: PermissionService
// Account domain defines what it needs for authorization
// ============================================================

type PermissionService interface {
	// Role Assignment
	AssignAccountAdminRole(ctx context.Context, accountID, userID string) error
	AssignEventManagerRole(ctx context.Context, accountID, userID string) error
	AssignTeamMemberRole(ctx context.Context, accountID, userID string) error

	// Role Removal
	RemoveRole(ctx context.Context, accountID, userID, role string) error

	// Permission Checks
	HasPermission(ctx context.Context, userID, domain, resource, action string) bool

	// Account Role Checks
	IsAccountAdmin(ctx context.Context, accountID, userID string) bool
	IsEventManager(ctx context.Context, accountID, userID string) bool
	IsTeamMember(ctx context.Context, accountID, userID string) bool
}