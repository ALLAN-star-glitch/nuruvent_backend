// internal/modules/certificate/domain/account_reader.go

package domain

import "context"

// AccountReader provides read-only access to account metadata.
// The certificate module uses this to resolve the issuer name
// (e.g., "Africana College of Professionals") when issuing
// certificates, so recipients see the organizer's brand rather
// than the platform name.
type AccountReader interface {
	// GetAccountName returns the display name of the account,
	// falling back to the canonical name if display_name is empty.
	GetAccountName(ctx context.Context, accountID string) (string, error)
}