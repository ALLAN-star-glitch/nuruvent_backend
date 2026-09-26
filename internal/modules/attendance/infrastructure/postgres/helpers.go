package postgres

// nullableString returns nil for an empty string, or a pointer to the
// value otherwise.
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