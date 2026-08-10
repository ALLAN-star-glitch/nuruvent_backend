// pkg/validation/validator.go
package validation

// Validator provides all validation methods
type Validator struct {
	Phone    Phone
	Password Password
	Email    Email
	UUID     UUID
	URL      URL
	String   String
}

// New creates a new validator instance
func New() *Validator {
	return &Validator{
		Phone:    Phone{},
		Password: Password{},
		Email:    Email{},
		UUID:     UUID{},
		URL:      URL{},
		String:   String{},
	}
}