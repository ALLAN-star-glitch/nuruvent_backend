package attendancedomain

import (
	"context"
	"time"
)

// AuthSessionIssuer mints an auth session for an already-identified
// user. Used by the join handler after redeeming a join token so
// guests land on the meeting page authenticated.
type AuthSessionIssuer interface {
	IssueSessionForUser(
		ctx context.Context,
		userID string,
		ttl time.Duration,
	) (token string, expiresAt time.Time, err error)
}