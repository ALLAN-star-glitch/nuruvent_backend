// internal/modules/auth/authdomain/domain.go

package authdomain

import domains "github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared"

// ============================================================
// RE-EXPORTS from internal/shared/domains
// ============================================================
//
// The canonical implementation lives in internal/shared/domains. This file
// re-exports it under the authdomain namespace.
//
// DOMAIN MODEL (post-revamp):
//
//   "platform"          — Nuruvent staff only
//   "account:<uuid>"    — every tenant account
//
// Teams are NOT authorization domains. Team membership is enforced as
// data (team_members) in the service layer, not through Casbin.
//
// If you need to add a new domain helper, add it to shared/domains and
// re-export it here.

// ---- Constants ----

const (
	// DomainPlatform is the static domain for Nuruvent staff.
	DomainPlatform = domains.DomainPlatform

	// AccountDomainPrefix is prepended to an account UUID to form its domain.
	AccountDomainPrefix = domains.AccountDomainPrefix

	// DomainDeferred signals that the domain cannot be resolved from the
	// request alone (e.g. slug-based lookups). The service layer MUST
	// perform the authorization check after loading the resource.
	DomainDeferred = "deferred"
)

// ---- Builders ----

var (
	// AccountDomain returns "account:<accountID>".
	AccountDomain = domains.AccountDomain
)

// ---- Checkers ----

var (
	IsAccountDomain  = domains.IsAccountDomain
	IsPlatformDomain = domains.IsPlatformDomain
	IsKnownDomain    = domains.IsKnownDomain
)

// ---- Extractors ----

var (
	ExtractAccountIDFromDomain = domains.ExtractAccountIDFromDomain
)

// ---- Parsers ----

var (
	ParseDomain = domains.ParseDomain
)