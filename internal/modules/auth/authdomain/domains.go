// internal/modules/auth/authdomain/domain.go

package authdomain

import "strings"

// ============================================================
// DOMAIN CONSTANTS
// ============================================================

const (
	// Platform domain
	DomainPlatform = "platform"

	// Team domain prefixes
	TeamDomainPrefixPersonal    = "personal:team:"
	TeamDomainPrefixInstitution = "institution:team:"
	TeamDomainPrefixAccount     = "account:"

	// Team type slugs (must match token claims)
	TeamTypePersonal    = "personal"
	TeamTypeInstitution = "institution"
)

// ============================================================
// DOMAIN BUILDERS
// ============================================================

// PersonalTeamDomain returns the personal team domain for a team.
// Format: "personal:team:{team_id}"
//
// ⚠️ The parameter is a TEAM ID, not a user ID. Every user has their own
// personal team (a real row in the `teams` table with its own UUID), and
// that team ID is what Casbin policies reference.
func PersonalTeamDomain(teamID string) string {
	if teamID == "" {
		return ""
	}
	return TeamDomainPrefixPersonal + teamID
}

// InstitutionTeamDomain returns the institution team domain for a team.
// Format: "institution:team:{team_id}"
//
// ⚠️ Note: this takes a TEAM ID, not an account ID. Institution teams belong
// to accounts (teams.account_id), but the domain is scoped to the team.
// The account domain (AccountDomain) is used for account-level fallback checks.
func InstitutionTeamDomain(teamID string) string {
	if teamID == "" {
		return ""
	}
	return TeamDomainPrefixInstitution + teamID
}

// AccountDomain returns the account domain for an account.
// Format: "account:{account_id}"
func AccountDomain(accountID string) string {
	if accountID == "" {
		return ""
	}
	return TeamDomainPrefixAccount + accountID
}

// BuildTeamDomain dispatches to the correct team domain builder based on team type.
// teamType must be "personal" or "institution"; empty defaults to "institution".
func BuildTeamDomain(teamType, teamID string) string {
	if teamID == "" {
		return ""
	}
	switch teamType {
	case TeamTypePersonal:
		return PersonalTeamDomain(teamID)
	case TeamTypeInstitution:
		return InstitutionTeamDomain(teamID)
	default:
		// Default to institution for safety when type is unknown
		return InstitutionTeamDomain(teamID)
	}
}

// ============================================================
// DOMAIN CHECKERS
// ============================================================

// IsPersonalTeamDomain checks if a domain is a personal team domain
func IsPersonalTeamDomain(domain string) bool {
	return strings.HasPrefix(domain, TeamDomainPrefixPersonal)
}

// IsInstitutionTeamDomain checks if a domain is an institution team domain
func IsInstitutionTeamDomain(domain string) bool {
	return strings.HasPrefix(domain, TeamDomainPrefixInstitution)
}

// IsAccountDomain checks if a domain is an account domain
func IsAccountDomain(domain string) bool {
	return strings.HasPrefix(domain, TeamDomainPrefixAccount)
}

// IsTeamDomain checks if a domain is a team domain (personal or institution)
func IsTeamDomain(domain string) bool {
	return IsPersonalTeamDomain(domain) || IsInstitutionTeamDomain(domain)
}

// IsPlatformDomain checks if a domain is the platform domain
func IsPlatformDomain(domain string) bool {
	return domain == DomainPlatform
}

// ============================================================
// EXTRACTION HELPERS
// ============================================================

// ExtractTeamID extracts the team identifier from a team domain.
//
// Semantics:
//   - "personal:team:{user_id}"      -> user_id    (which IS the personal team id)
//   - "institution:team:{team_id}"   -> team_id
//   - "account:{account_id}"         -> account_id (kept for legacy compat)
//   - "platform"                     -> ""
func ExtractTeamID(domain string) string {
	switch {
	case IsPersonalTeamDomain(domain):
		return strings.TrimPrefix(domain, TeamDomainPrefixPersonal)
	case IsInstitutionTeamDomain(domain):
		return strings.TrimPrefix(domain, TeamDomainPrefixInstitution)
	case IsAccountDomain(domain):
		return strings.TrimPrefix(domain, TeamDomainPrefixAccount)
	default:
		return ""
	}
}

// ExtractAccountIDFromDomain extracts the account identifier carried by an
// account-scoped domain.
//
// Semantics:
//   - "account:{account_id}"         -> account_id
//   - "institution:team:{team_id}"   -> "" (team domain does NOT carry account id)
//   - "personal:team:{user_id}"      -> "" (personal scope has no account)
//
// ⚠️ This previously returned the team ID for institution domains — that was
// incorrect. If you need the account for an institution team, resolve it via
// the team service (teams.account_id), not by string-parsing the domain.
func ExtractAccountIDFromDomain(domain string) string {
	if IsAccountDomain(domain) {
		return strings.TrimPrefix(domain, TeamDomainPrefixAccount)
	}
	return ""
}

// ExtractTeamType extracts team type from a team domain.
// Returns: "personal", "institution", or "" (for non-team domains)
func ExtractTeamType(domain string) string {
	switch {
	case IsPersonalTeamDomain(domain):
		return TeamTypePersonal
	case IsInstitutionTeamDomain(domain):
		return TeamTypeInstitution
	default:
		return ""
	}
}

// ============================================================
// CONTEXT HELPERS (token claims -> domain)
// ============================================================

// TeamDomainFromClaims builds a team domain from raw token claims.
// teamType should be "personal" or "institution"; teamID is the token's team_id.
// Returns "" if teamID is empty.
//
// Use this in auth middleware and authz middleware to convert JWT claims into
// the domain string the Casbin checker expects.
func TeamDomainFromClaims(teamType, teamID string) string {
	return BuildTeamDomain(teamType, teamID)
}

// ResolveEffectiveDomains returns the 2-tier lookup chain for permission checks:
//
//	Tier 1: team domain   (personal:team:X or institution:team:X)
//	Tier 2: account domain (account:Y) — omitted if accountID is empty
//
// The returned slice preserves order: [teamDomain, accountDomain].
// Either element may be "" if its source ID is empty.
//
// Typical usage:
//
//	for _, domain := range ResolveEffectiveDomains(teamType, teamID, accountID) {
//	    if domain == "" { continue }
//	    allowed, _ := checker.HasPermission(ctx, userID, domain, resource, action)
//	    if allowed { return true }
//	}
//	return false
func ResolveEffectiveDomains(teamType, teamID, accountID string) []string {
	domains := make([]string, 0, 2)

	if teamDomain := BuildTeamDomain(teamType, teamID); teamDomain != "" {
		domains = append(domains, teamDomain)
	}

	if accountDomain := AccountDomain(accountID); accountDomain != "" {
		domains = append(domains, accountDomain)
	}

	return domains
}