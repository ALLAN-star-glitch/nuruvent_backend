// internal/modules/auth/authdomain/domains.go

package authdomain

import "strings"

// Domain constants
const (
    DomainPlatform = "platform"

    // Team domain prefixes
    TeamDomainPrefixPersonal    = "personal:team:"
    TeamDomainPrefixInstitution = "institution:team:"

    // Account domain prefix
    AccountDomainPrefix = "account:"
)

// ============================================================
// PERSONAL TEAM DOMAINS
// ============================================================

// PersonalTeamDomain returns the personal team domain for a user
// Format: "personal:team:{user_id}"
func PersonalTeamDomain(userID string) string {
    if userID == "" {
        return ""
    }
    return TeamDomainPrefixPersonal + userID
}

// IsPersonalTeamDomain checks if a domain is a personal team domain
func IsPersonalTeamDomain(domain string) bool {
    return strings.HasPrefix(domain, TeamDomainPrefixPersonal)
}

// ============================================================
// INSTITUTION TEAM DOMAINS
// ============================================================

// InstitutionTeamDomain returns the institution team domain
// Format: "institution:team:{institution_id}"
func InstitutionTeamDomain(institutionID string) string {
    if institutionID == "" {
        return ""
    }
    return TeamDomainPrefixInstitution + institutionID
}

// IsInstitutionTeamDomain checks if a domain is an institution team domain
func IsInstitutionTeamDomain(domain string) bool {
    return strings.HasPrefix(domain, TeamDomainPrefixInstitution)
}

// ============================================================
// ACCOUNT DOMAINS (NEW)
// ============================================================

// AccountDomain returns the account domain for an account
// Format: "account:{account_id}"
func AccountDomain(accountID string) string {
    if accountID == "" {
        return ""
    }
    return AccountDomainPrefix + accountID
}

// IsAccountDomain checks if a domain is an account domain
func IsAccountDomain(domain string) bool {
    return strings.HasPrefix(domain, AccountDomainPrefix)
}

// ExtractAccountID extracts account ID from an account domain
func ExtractAccountID(domain string) string {
    if IsAccountDomain(domain) {
        return strings.TrimPrefix(domain, AccountDomainPrefix)
    }
    return ""
}

// ============================================================
// TEAM DOMAINS
// ============================================================

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