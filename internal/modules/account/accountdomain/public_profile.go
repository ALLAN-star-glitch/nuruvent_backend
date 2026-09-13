package accountdomain

// PublicProfile is the public-facing projection of a user.
// It intentionally omits email, phone, and verification flags.
type PublicProfile struct {
	ID          string
	DisplayName string
	AvatarURL   string
	Bio         string
	Location    string
	Website     string
	SocialLinks map[string]string
}