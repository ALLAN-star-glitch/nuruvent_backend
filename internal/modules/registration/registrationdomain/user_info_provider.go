// internal/modules/registration/registrationdomain/user_info_provider.go

package registrationdomain

import "context"

// UserInfoProvider resolves a user's display name and email by ID.
//
// Backed by a cross-module adapter that delegates to auth's
// GetUserByID. The registration module never imports auth directly.
type UserInfoProvider interface {
	GetUserInfo(ctx context.Context, userID string) (displayName, email string, err error)
	GetUsername(ctx context.Context, userID string) (string, error)
	GetPhone(ctx context.Context, userID string) (string, error)


	// FindOrCreateGuestByEmail returns the user ID for the given
	// email, creating a lightweight guest user if none exists.
	//
	// Used to ensure every registration has a user_id, which is
	// required by the Zoom SDK and the attendance pipeline.
	FindOrCreateGuestByEmail(
		ctx context.Context,
		email string,
		name string,
		phone string,
	) (string, error)
}