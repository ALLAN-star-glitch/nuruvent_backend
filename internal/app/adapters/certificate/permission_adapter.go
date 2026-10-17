// internal/app/adapters/certificate/permission_adapter.go

package certificate

import (
	"context"

	authDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
)

type PermissionAdapter struct {
	perm authDomain.PermissionChecker
}

func NewPermissionAdapter(perm authDomain.PermissionChecker) *PermissionAdapter {
	return &PermissionAdapter{perm: perm}
}

func (a *PermissionAdapter) CanManageTemplates(ctx context.Context, userID, accountID string) (bool, error) {
	return a.perm.HasAnyPermission(ctx, userID, authDomain.AccountDomain(accountID),
		authDomain.ResourceCertificate.String(),
		authDomain.ActionCreate.String(),
		authDomain.ActionUpdate.String(),
		authDomain.ActionDelete.String(),
	)
}

func (a *PermissionAdapter) CanIssue(ctx context.Context, userID, accountID string) (bool, error) {
	return a.perm.HasPermission(ctx, userID, authDomain.AccountDomain(accountID),
		authDomain.ResourceCertificate.String(),
		authDomain.ActionIssue.String(),
	)
}

func (a *PermissionAdapter) CanReadAll(ctx context.Context, userID, accountID string) (bool, error) {
	return a.perm.HasPermission(ctx, userID, authDomain.AccountDomain(accountID),
		authDomain.ResourceCertificate.String(),
		authDomain.ActionRead.String(),
	)
}

func (a *PermissionAdapter) CanReadOwn(ctx context.Context, userID, accountID string) (bool, error) {
	return a.perm.HasAnyPermission(ctx, userID, authDomain.AccountDomain(accountID),
		authDomain.ResourceCertificate.String(),
		authDomain.ActionReadOwn.String(),
		authDomain.ActionRead.String(),
	)
}

func (a *PermissionAdapter) CanDownload(ctx context.Context, userID, accountID string) (bool, error) {
	return a.perm.HasAnyPermission(ctx, userID, authDomain.AccountDomain(accountID),
		authDomain.ResourceCertificate.String(),
		authDomain.ActionDownload.String(),
		authDomain.ActionRead.String(),
	)
}

func (a *PermissionAdapter) CanRevoke(ctx context.Context, userID, accountID string) (bool, error) {
	return a.perm.HasAnyPermission(ctx, userID, authDomain.AccountDomain(accountID),
		authDomain.ResourceCertificate.String(),
		authDomain.ActionUpdate.String(),
		authDomain.ActionDelete.String(),
	)
}

// CanManageSystemTemplates checks the PLATFORM domain, where super_admin
// and admin roles live. Account-level roles (account_admin, trainer,
// learner) have no grants in the platform domain, so this correctly
// returns false for them.
func (a *PermissionAdapter) CanManageSystemTemplates(ctx context.Context, userID string) (bool, error) {
	return a.perm.HasAnyPermission(
		ctx,
		userID,
		authDomain.DomainPlatform, // <- literal "platform"
		authDomain.ResourceCertificate.String(),
		authDomain.ActionManage.String(),
		authDomain.ActionCreate.String(),
	)
}

var _ domain.AuthorizationChecker = (*PermissionAdapter)(nil)