
package validation

// Validator provides all validation methods
type Validator struct {
	Phone    Phone
	Password Password
	Email    Email
	UUID     UUID
	URL      URL
	String   String
	Sanitize Sanitize
}

// New creates a new validator instance
// New creates a new validator instance
func New() *Validator {
	return &Validator{
		Phone:    Phone{DefaultRegion: "KE"},
		Password: Password{},
		Email:    Email{},
		UUID:     UUID{},
		URL:      URL{},
		String:   String{},
		Sanitize: Sanitize{},
	}
}