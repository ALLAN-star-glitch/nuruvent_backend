package accountdomain

// UserInfo is the read-only projection of a user, exposed to other modules
// for display purposes (event creator, audit logs, etc.).
//
// This is NOT the full user entity. It carries only the fields external
// consumers need and intentionally omits internal fields (password hash,
// internal flags, timestamps).
type UserInfo struct {
	ID          string
	Name        string
	DisplayName string
	Email       string
	Phone       string
	AvatarURL   string
	Bio         string
	Location    string
	Website     string
	SocialLinks map[string]string
}