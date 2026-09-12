// internal/modules/team/service/domain.go

package service

import domains "github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared"

// ============================================================
// RE-EXPORTS from internal/shared/domains
// ============================================================
//
// The canonical implementation lives in internal/shared/domains. This file
// re-exports it under the team service namespace so existing call sites
// (team service, permission helpers, etc.) continue to work unchanged.
//
// The team module must not import the auth module directly. Both modules
// depend on shared/domains instead — that's the only shared contract.
//
// DOMAIN MODEL (post-revamp):
//
//   "platform"        — Nuruvent staff only
//   "account:<uuid>"  — every tenant account
//
// Teams are NOT authorization domains. A team's authorization is checked
// against its parent ACCOUNT domain ("account:<team.account_id>"). Team
// membership is enforced as data (team_members) in the service layer, not
// through Casbin.
//
// REMOVED TYPES (migration reference):
//
//   - Domain (interface)
//   - PersonalTeamDomain    (type + constructor)
//   - InstitutionTeamDomain (type + constructor)
//   - AccountDomain         (type + constructor)
//   - NewTeamDomain / NewDomainFromTeam / NewDomainFromTeamID
//
// Callers should:
//
//   - Use AccountDomain(team.AccountID) to obtain the authz domain for
//     a team.
//   - Use team_members table queries for team visibility checks.

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