// internal/modules/auth/authdomain/username.go

package authdomain

import (
	"fmt"
	"regexp"
	"strings"
)

var nonAlnum = regexp.MustCompile(`[^a-z0-9]`)

// GenerateUsername derives a base username from an email address.
// Lowercase, alphanumerics only. Returns "user" if the local part
// contains no usable characters.
func GenerateUsername(email string) string {
	local := strings.ToLower(strings.TrimSpace(email))
	if at := strings.Index(local, "@"); at > 0 {
		local = local[:at]
	}
	local = nonAlnum.ReplaceAllString(local, "")
	if local == "" {
		local = "user"
	}
	return local
}

// UsernameCandidates yields usernames to try, in order. The caller
// checks each against the DB until one is free.
func UsernameCandidates(email string) []string {
	base := GenerateUsername(email)
	out := make([]string, 0, 101)
	out = append(out, base)
	for i := 2; i <= 100; i++ {
		out = append(out, fmt.Sprintf("%s%d", base, i))
	}
	return out
}