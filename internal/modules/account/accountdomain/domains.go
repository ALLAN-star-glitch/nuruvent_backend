// internal/modules/account/accountdomain/domains.go

package accountdomain

import domains "github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared"

// ============================================================
// PUBLIC FACADE — account module's domain-string vocabulary
// ============================================================
//
// This file re-exports the shared domain builders, checkers, and extractors
// under the account domain namespace.
//
// WHY THIS FILE EXISTS:
//
//   - The account module's `accountdomain` package is its public API.
//     Other modules import from here, not from shared/domains directly.
//
//   - When this module is eventually extracted into a separate service,
//     this facade becomes the boundary. It can be reimplemented to build
//     domain strings from whatever source the extracted service uses —
//     without touching any call site inside the account module.
//
//   - Every cross-module consumer of domain strings goes through this file.
//
// DOMAIN MODEL (post-revamp):
//
//   "platform"        — Nuruvent staff only
//   "account:<uuid>"  — every tenant account
//
// Teams are NOT authorization domains. Team membership is enforced as
// data (team_members) in the service layer, not through Casbin.

// ---- Constants ----

const (
	// DomainPlatform is the static domain for Nuruvent staff.
	DomainPlatform = domains.DomainPlatform

	// AccountDomainPrefix is prepended to an account UUID to form its domain.
	AccountDomainPrefix = domains.AccountDomainPrefix
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