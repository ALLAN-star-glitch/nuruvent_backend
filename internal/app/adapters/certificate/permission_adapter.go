// internal/app/adapters/certificate/permission_adapter.go

package certificate

import (
	"context"

	authDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
)

// PermissionAdapter bridges auth's PermissionChecker to the certificate
// module's domain.AuthorizationChecker port.
//
// The adapter translates the certificate module's (userID, accountID)
// call shape into auth's (userID, domain, resource, action) shape, and
// maps each certificate concern to the right permission check.
type PermissionAdapter struct {
	perm authDomain.PermissionChecker
}

// NewPermissionAdapter constructs a PermissionAdapter.
func NewPermissionAdapter(perm authDomain.PermissionChecker) *PermissionAdapter {
	return &PermissionAdapter{perm: perm}
}

// CanManageTemplates maps to certificate:create / update / delete.
func (a *PermissionAdapter) CanManageTemplates(ctx context.Context, userID, accountID string) (bool, error) {
	return a.perm.HasAnyPermission(ctx, userID, authDomain.AccountDomain(accountID),
		authDomain.ResourceCertificate.String(),
		authDomain.ActionCreate.String(),
		authDomain.ActionUpdate.String(),
		authDomain.ActionDelete.String(),
	)
}

// CanIssue maps to certificate:issue.
func (a *PermissionAdapter) CanIssue(ctx context.Context, userID, accountID string) (bool, error) {
	return a.perm.HasPermission(ctx, userID, authDomain.AccountDomain(accountID),
		authDomain.ResourceCertificate.String(),
		authDomain.ActionIssue.String(),
	)
}

// CanReadAll maps to certificate:read (account-wide).
func (a *PermissionAdapter) CanReadAll(ctx context.Context, userID, accountID string) (bool, error) {
	return a.perm.HasPermission(ctx, userID, authDomain.AccountDomain(accountID),
		authDomain.ResourceCertificate.String(),
		authDomain.ActionRead.String(),
	)
}

// CanReadOwn maps to certificate:read_own.
//
// NOTE: this action is not currently in DefaultAccountPermissions for
// any role. Learners get certificate:read (see authdomain). If the
// policy later adds certificate:read_own, this adapter will pick it up
// automatically. For now, fall back to certificate:read so a learner
// can list their own certificates.
func (a *PermissionAdapter) CanReadOwn(ctx context.Context, userID, accountID string) (bool, error) {
	return a.perm.HasAnyPermission(ctx, userID, authDomain.AccountDomain(accountID),
		authDomain.ResourceCertificate.String(),
		authDomain.ActionReadOwn.String(),
		authDomain.ActionRead.String(),
	)
}

// CanDownload maps to certificate:download.
func (a *PermissionAdapter) CanDownload(ctx context.Context, userID, accountID string) (bool, error) {
	return a.perm.HasAnyPermission(ctx, userID, authDomain.AccountDomain(accountID),
		authDomain.ResourceCertificate.String(),
		authDomain.ActionDownload.String(),
		authDomain.ActionRead.String(),
	)
}

// CanRevoke maps to certificate:update (revocation is an update to a
// certificate's state).
func (a *PermissionAdapter) CanRevoke(ctx context.Context, userID, accountID string) (bool, error) {
	return a.perm.HasAnyPermission(ctx, userID, authDomain.AccountDomain(accountID),
		authDomain.ResourceCertificate.String(),
		authDomain.ActionUpdate.String(),
		authDomain.ActionDelete.String(),
	)
}

// compile-time assertion
var _ domain.AuthorizationChecker = (*PermissionAdapter)(nil)