// internal/modules/video/infrastructure/postgres/helpers.go

package postgres

import "time"

// nullableString returns nil for an empty string, or a pointer to the
// value otherwise. Used for nullable text columns.
func nullableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// derefString returns the value behind a pointer, or "" if nil.
func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// durationFromSeconds converts an int stored in the DB back to a
// time.Duration.
func durationFromSeconds(sec int) time.Duration {
	return time.Duration(sec) * time.Second
}