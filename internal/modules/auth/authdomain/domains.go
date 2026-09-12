// internal/modules/auth/authdomain/domain.go

package authdomain

import domains "github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared"

// ============================================================
// RE-EXPORTS from internal/shared/domains
// ============================================================
//
// The canonical implementation lives in internal/shared/domains. This file
// re-exports it under the authdomain namespace so existing callers
// (auth middleware, authorization middleware, profile, etc.) continue to
// work unchanged.
//
// If you need to add a new domain helper, add it to shared/domains and
// re-export it here.

// ---- Constants ----

const (
	DomainPlatform              = domains.DomainPlatform
	TeamDomainPrefixPersonal    = domains.TeamDomainPrefixPersonal
	TeamDomainPrefixInstitution = domains.TeamDomainPrefixInstitution
	TeamDomainPrefixAccount     = domains.TeamDomainPrefixAccount
	TeamTypePersonal            = domains.TeamTypePersonal
	TeamTypeInstitution         = domains.TeamTypeInstitution
	// DomainDeferred signals that the domain cannot be resolved from the
    // request alone (e.g. slug-based lookups). The service layer MUST
    // perform the authorization check after loading the resource.
    DomainDeferred = "deferred"
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