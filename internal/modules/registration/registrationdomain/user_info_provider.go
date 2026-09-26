// internal/modules/registration/registrationdomain/user_info_provider.go

package registrationdomain

import "context"

// UserInfoProvider resolves a user's display name and email by ID.
//
// Backed by a cross-module adapter that delegates to auth's
// GetUserByID. The registration module never imports auth directly.
type UserInfoProvider interface {
	GetUserInfo(ctx context.Context, userID string) (displayName, email string, err error)
}