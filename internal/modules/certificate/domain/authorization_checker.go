// internal/modules/certificate/domain/authorization_checker.go

package domain

import "context"

// AuthorizationChecker answers "can this user do this?" for the
// certificate module. It is deliberately narrow — one method per
// certificate concern — so the certificate module never leaks the
// shape of the underlying auth system (Casbin, RBAC, ABAC, whatever).
//
// Callers pass a domain string in the auth module's format
// ("account:<uuid>"), because the certificate module already knows
// the account ID it is operating in. The certificate module does NOT
// interpret the string — it just forwards it.
//
// An adapter wraps auth's PermissionChecker and implements this
// interface.
type AuthorizationChecker interface {
	// CanManageTemplates reports whether userID can create, update, or
	// delete templates in the given account domain.
	CanManageTemplates(ctx context.Context, userID, accountDomain string) (bool, error)

	// CanIssue reports whether userID can issue certificates in the
	// given account domain.
	CanIssue(ctx context.Context, userID, accountDomain string) (bool, error)

	// CanReadAll reports whether userID can list all certificates in
	// the given account domain.
	CanReadAll(ctx context.Context, userID, accountDomain string) (bool, error)

	// CanReadOwn reports whether userID can list their own certificates.
	// This is not always account-scoped — a learner may read their own
	// certificates across accounts they belong to.
	CanReadOwn(ctx context.Context, userID, accountDomain string) (bool, error)

	// CanDownload reports whether userID can download a specific
	// certificate.
	CanDownload(ctx context.Context, userID, accountDomain string) (bool, error)

	// CanRevoke reports whether userID can revoke a certificate.
	CanRevoke(ctx context.Context, userID, accountDomain string) (bool, error)


	// CanManageSystemTemplates reports whether userID can create or manage
	// system-wide templates (only available to super_admins/admins).
	CanManageSystemTemplates(ctx context.Context, userID string) (bool, error)
}