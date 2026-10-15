// internal/modules/certificate/domain/metadata.go

package domain

import "time"

// Metadata carries free-form key-value pairs that the template renders.
//
// The certificate module never interprets these values — it only
// forwards them to the renderer. The keys below are a published
// convention so templates can rely on standard names regardless of
// whether the caller is the events module or the courses module.
type Metadata map[string]string

// Convention keys. Callers are encouraged (not required) to set these.
const (
	MetadataKeyContextTitle    = "context_title"
	MetadataKeyContextDate     = "context_date"
	MetadataKeyContextDuration = "context_duration"
	MetadataKeyIssuerName      = "issuer_name"
	MetadataKeyRecipientName   = "recipient_name"
)

// Get returns the value for key, or "" if absent.
func (m Metadata) Get(key string) string { return m[key] }

// Has reports whether key is present.
func (m Metadata) Has(key string) bool {
	_, ok := m[key]
	return ok
}

// Clone returns a defensive copy.
func (m Metadata) Clone() Metadata {
	if m == nil {
		return nil
	}
	out := make(Metadata, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

// WithContextDate sets the conventional context_date key from a time.
func (m Metadata) WithContextDate(t time.Time) Metadata {
	if m == nil {
		m = Metadata{}
	}
	m[MetadataKeyContextDate] = t.UTC().Format(time.RFC3339)
	return m
}