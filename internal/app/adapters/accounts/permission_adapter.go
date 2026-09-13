package accounts

import (
	"context"

	accountdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/accountdomain"
	authdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
)

// PermissionAdapter bridges authdomain.PermissionChecker to
// accountdomain.PermissionChecker.
//
// The concrete implementation (authorization.PermissionChecker) already
// satisfies both interfaces structurally — no method mapping is needed.
// This adapter exists purely to make Wire's interface wiring explicit
// and to give each module a named dependency it can depend on.
//
// Why an adapter instead of wire.Bind?
//   - Keeps all cross-module interface translations in one place
//     (app/adapters/accounts/*), rather than scattering wire.Bind
//     directives across provider sets.
//   - Makes the dependency direction visible: the account module
//     consumes the auth module's permission checker.
type PermissionAdapter struct {
	checker authdomain.PermissionChecker
}

// NewPermissionAdapter wraps an authdomain.PermissionChecker and exposes it
// as an accountdomain.PermissionChecker.
func NewPermissionAdapter(checker authdomain.PermissionChecker) accountdomain.PermissionChecker {
	return &PermissionAdapter{checker: checker}
}

// ============================================================
// accountdomain.PermissionChecker
// ============================================================

func (a *PermissionAdapter) HasPermission(ctx context.Context, userID, domain, resource, action string) (bool, error) {
	return a.checker.HasPermission(ctx, userID, domain, resource, action)
}

func (a *PermissionAdapter) HasAnyPermission(ctx context.Context, userID, domain, resource string, actions ...string) (bool, error) {
	return a.checker.HasAnyPermission(ctx, userID, domain, resource, actions...)
}

func (a *PermissionAdapter) HasAllPermissions(ctx context.Context, userID, domain, resource string, actions ...string) (bool, error) {
	return a.checker.HasAllPermissions(ctx, userID, domain, resource, actions...)
}

func (a *PermissionAdapter) CanManageAccount(ctx context.Context, userID, domain string) (bool, error) {
	return a.checker.CanManageAccount(ctx, userID, domain)
}

func (a *PermissionAdapter) CanManageAccountMembers(ctx context.Context, userID, domain string) (bool, error) {
	return a.checker.CanManageAccount(ctx, userID, domain)
}

func (a *PermissionAdapter) CanViewAccount(ctx context.Context, userID, domain string) (bool, error) {
	return a.checker.CanViewAccount(ctx, userID, domain)
}