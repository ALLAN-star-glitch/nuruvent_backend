// internal/modules/profile/domain/domain.go

package domain

import "strings"

// Domain constants
const (
	// Team domain prefixes
	TeamDomainPrefixPersonal    = "personal:team:"
	TeamDomainPrefixInstitution = "institution:team:"
	TeamDomainPrefixAccount     = "account:"
)

// ============================================================
// TEAM DOMAINS
// ============================================================

// PersonalTeamDomain returns the personal team domain for a user
// Format: "personal:team:{user_id}"
func PersonalTeamDomain(userID string) string {
	if userID == "" {
		return ""
	}
	return TeamDomainPrefixPersonal + userID
}

// InstitutionTeamDomain returns the institution team domain for an account
// Format: "institution:team:{account_id}"
func InstitutionTeamDomain(accountID string) string {
	if accountID == "" {
		return ""
	}
	return TeamDomainPrefixInstitution + accountID
}

// AccountDomain returns the account domain for an account
// Format: "account:{account_id}"
func AccountDomain(accountID string) string {
	if accountID == "" {
		return ""
	}
	return TeamDomainPrefixAccount + accountID
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

// ============================================================
// EXTRACTION HELPERS
// ============================================================

// ExtractTeamID extracts team ID from a team domain
func ExtractTeamID(domain string) string {
	if IsPersonalTeamDomain(domain) {
		return strings.TrimPrefix(domain, TeamDomainPrefixPersonal)
	}
	if IsInstitutionTeamDomain(domain) {
		return strings.TrimPrefix(domain, TeamDomainPrefixInstitution)
	}
	if IsAccountDomain(domain) {
		return strings.TrimPrefix(domain, TeamDomainPrefixAccount)
	}
	return ""
}

// ExtractAccountIDFromDomain extracts account ID from a domain
// Works with: "institution:team:{account_id}", "account:{account_id}", "personal:team:{user_id}"
func ExtractAccountIDFromDomain(domain string) string {
	if IsInstitutionTeamDomain(domain) {
		return ExtractTeamID(domain)
	}
	if IsAccountDomain(domain) {
		return ExtractTeamID(domain)
	}
	// For personal teams, the account ID is the user ID
	if IsPersonalTeamDomain(domain) {
		return ExtractTeamID(domain)
	}
	return ""
}

// ExtractTeamType extracts team type from a team domain
// Returns: "personal", "institution", or empty string
func ExtractTeamType(domain string) string {
	if IsPersonalTeamDomain(domain) {
		return "personal"
	}
	if IsInstitutionTeamDomain(domain) {
		return "institution"
	}
	return ""
}