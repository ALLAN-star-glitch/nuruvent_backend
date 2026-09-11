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
//     without touching any callsite inside the account module.
//
//   - Every cross-module consumer of domain strings goes through this file.
//
// If you're reading this and thinking "this looks like duplication" —
// it's intentional. It's the seam.

// ---- Constants ----

const (
	DomainPlatform              = domains.DomainPlatform
	TeamDomainPrefixPersonal    = domains.TeamDomainPrefixPersonal
	TeamDomainPrefixInstitution = domains.TeamDomainPrefixInstitution
	TeamDomainPrefixAccount     = domains.TeamDomainPrefixAccount
	TeamTypePersonal            = domains.TeamTypePersonal
	TeamTypeInstitution         = domains.TeamTypeInstitution
)

// ---- Builders ----

var (
	PersonalTeamDomain    = domains.PersonalTeamDomain
	InstitutionTeamDomain = domains.InstitutionTeamDomain
	AccountDomain         = domains.AccountDomain
	BuildTeamDomain       = domains.BuildTeamDomain
	TeamDomainFromClaims  = domains.TeamDomainFromClaims
)

// ---- Checkers ----

var (
	IsPersonalTeamDomain    = domains.IsPersonalTeamDomain
	IsInstitutionTeamDomain = domains.IsInstitutionTeamDomain
	IsAccountDomain         = domains.IsAccountDomain
	IsTeamDomain            = domains.IsTeamDomain
	IsPlatformDomain        = domains.IsPlatformDomain
)

// ---- Extractors ----

var (
	ExtractTeamID              = domains.ExtractTeamID
	ExtractAccountIDFromDomain = domains.ExtractAccountIDFromDomain
	ExtractTeamType            = domains.ExtractTeamType
)

// ---- Higher-level helpers ----

var ResolveEffectiveDomains = domains.ResolveEffectiveDomains