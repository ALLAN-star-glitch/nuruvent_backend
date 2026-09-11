package accountdomain

import "time"

// User is the core user entity owned by the account module.
// (Or however your system models users — adapt to what you have.)
type User struct {
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
	Slug        string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}