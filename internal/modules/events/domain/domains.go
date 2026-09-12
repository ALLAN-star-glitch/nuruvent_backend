// internal/modules/events/domain/domains.go

package domain

import domains "github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared"

// ============================================================
// RE-EXPORTS from internal/shared/domains
// ============================================================
//
// The canonical implementation lives in internal/shared/domains. This file
// re-exports it under the events domain namespace so existing call sites
// (event service, permission helpers, etc.) continue to work unchanged.
//
// The events module must not import the auth module directly. Both modules
// depend on shared/domains instead — that's the only shared contract.
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