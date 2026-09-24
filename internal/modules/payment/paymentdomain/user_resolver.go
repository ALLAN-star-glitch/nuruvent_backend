package paymentdomain

import "context"

// UserEmailResolver resolves a user ID to an email address. It is used
// by notification adapters that need to send to authenticated users
// without importing the account module directly.
type UserEmailResolver interface {
	ResolveEmail(ctx context.Context, userID string) (string, error)
}