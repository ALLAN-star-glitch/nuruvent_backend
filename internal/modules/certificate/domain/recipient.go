// internal/modules/certificate/domain/recipient.go

package domain

import "strings"

// Recipient identifies who receives a certificate.
//
// Name and Email are snapshotted onto the certificate at issue time.
// UserID is optional — guest recipients may not have an account.
type Recipient struct {
	UserID string
	Name   string
	Email  string
}

// Validate ensures the recipient carries the minimum required data.
func (r Recipient) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return ErrRecipientNameRequired
	}
	if strings.TrimSpace(r.Email) == "" {
		return ErrRecipientEmailRequired
	}
	return nil
}

// normalizedEmail returns the lowercased, trimmed email. Used for
// identity comparison and snapshotting.
func (r Recipient) normalizedEmail() string {
	return strings.ToLower(strings.TrimSpace(r.Email))
}