// internal/shared/domains/domains.go

package domains

import (
	"fmt"
	"strings"
)

// ============================================================
// SHARED DOMAIN VOCABULARY
// ============================================================
//
// Domain strings are the shared vocabulary that Casbin policies use to
// identify scopes.
//
// DOMAIN MODEL (post-revamp):
//
//	"platform"          — Nuruvent staff only (super_admin, admin, guest)
//	"account:<uuid>"    — every tenant account
//
// Teams are NOT authorization domains. Team membership is enforced as
// data (team_members) in the service layer, not through Casbin.
//
// This package is the single source of truth for building, checking, and
// parsing domain strings. Auth, accounts, events, and any future module
// that performs permission checks should depend on this package.
//
// Every function here is pure: no I/O, no context, no side effects.

// ============================================================
// DOMAIN CONSTANTS
// ============================================================

const (
	// DomainPlatform is the literal token for platform-wide scope.
	DomainPlatform = "platform"

	// AccountDomainPrefix is prepended to an account UUID to form its domain.
	AccountDomainPrefix = "account:"
)

// ============================================================
// DOMAIN BUILDERS
// ============================================================

// AccountDomain returns "account:<accountID>". Returns "" for empty input.
func AccountDomain(accountID string) string {
	if accountID == "" {
		return ""
	}
	return AccountDomainPrefix + accountID
}

// ============================================================
// DOMAIN CHECKERS
// ============================================================

// IsAccountDomain reports whether d is an account domain.
func IsAccountDomain(d string) bool {
	return strings.HasPrefix(d, AccountDomainPrefix)
}

// IsPlatformDomain reports whether d is the platform domain.
func IsPlatformDomain(d string) bool {
	return d == DomainPlatform
}

// IsKnownDomain reports whether d is one of the recognized domain shapes.
// Use in guards to reject malformed domains early.
func IsKnownDomain(d string) bool {
	return IsPlatformDomain(d) || IsAccountDomain(d)
}

// ============================================================
// EXTRACTION HELPERS
// ============================================================

// ExtractAccountIDFromDomain returns the account UUID from "account:<uuid>",
// or "" if d is not an account domain.
func ExtractAccountIDFromDomain(d string) string {
	if !IsAccountDomain(d) {
		return ""
	}
	return strings.TrimPrefix(d, AccountDomainPrefix)
}

// ============================================================
// PARSERS
// ============================================================

// ParseDomain splits a domain into (kind, id).
//
//	"platform"        -> ("platform", "", nil)
//	"account:<uuid>"  -> ("account", "<uuid>", nil)
//	anything else     -> ("", "", error)
func ParseDomain(d string) (kind, id string, err error) {
	switch {
	case IsPlatformDomain(d):
		return "platform", "", nil
	case IsAccountDomain(d):
		return "account", ExtractAccountIDFromDomain(d), nil
	default:
		return "", "", fmt.Errorf("unrecognized domain: %q", d)
	}
}