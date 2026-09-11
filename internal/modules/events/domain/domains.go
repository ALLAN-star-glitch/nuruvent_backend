// internal/modules/events/domain/domains.go

package domain

import domains "github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared"

// ============================================================
// RE-EXPORTS from internal/shared/domains
// ============================================================
//
// The canonical implementation lives in internal/shared/domains. This file
// re-exports it under the events domain namespace so existing callsites
// (event service, permission helpers, etc.) continue to work unchanged.
//
// The events module must not import the auth module directly. Both modules
// depend on shared/domains instead — that's the only shared contract.

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