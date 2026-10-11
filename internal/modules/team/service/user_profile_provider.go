package service

import "context"

// UserProfile is the identity projection the team module needs for
// enriching member rosters. It carries only display fields — no
// credentials, no account state.
type UserProfile struct {
	ID          string
	Name        string
	DisplayName string
	Email       string
	AvatarURL   string
}

// UserProfileProvider is the team module's outbound port for resolving
// user display identities in batch.
//
// Avatars and profile fields live in the account module (see
// accountdomain.UserInfo), not on the auth user record. The team
// module therefore reads them through this port rather than through
// the auth port.
type UserProfileProvider interface {
	// GetUserProfilesByIDs returns a map of userID → UserProfile for
	// the given IDs. Users that don't exist are omitted.
	//
	// Batch-oriented: callers pass every user ID they need in one call.
	GetUserProfilesByIDs(
		ctx context.Context,
		userIDs []string,
	) (map[string]*UserProfile, error)
}