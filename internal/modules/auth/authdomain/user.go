// internal/modules/auth/authdomain/user.go

package authdomain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// User entity (formerly Account)
type User struct {
	ID             string
	Slug           string
	Name           string
	DisplayName    string
	Email          string
	PasswordHash   string
	Phone          string
	AccountTypeID  string
	ProfessionalTypeID *string
	InstitutionID  *string
	
	// Verification fields
	EmailVerified     bool
	EmailVerifiedAt   *time.Time
	PhoneVerified     bool
	PhoneVerifiedAt   *time.Time
	IdentityVerified  bool
	IdentityVerifiedAt *time.Time
	
	// KYC fields
	KYCStatus          string  // pending, submitted, verified, rejected, not_required
	KYCSubmittedAt     *time.Time
	KYCVerifiedAt      *time.Time
	KYCRejectedAt      *time.Time
	KYCRejectionReason *string
	
	// Document fields
	IDDocument     *string
	SelfieDocument *string
	AddressProof   *string
	
	// Status
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// NewUser creates a validated user
func NewUser(email, passwordHash, name, phone, accountTypeID string) (*User, error) {
	if email == "" {
		return nil, errors.New("email is required")
	}
	if passwordHash == "" {
		return nil, errors.New("password is required")
	}
	if name == "" {
		return nil, errors.New("name is required")
	}
	if accountTypeID == "" {
		return nil, errors.New("account type is required")
	}

	now := time.Now()
	return &User{
		ID:             uuid.New().String(),
		Slug:           slugify(name),
		Name:           name,
		DisplayName:    name,
		Email:          email,
		PasswordHash:   passwordHash,
		Phone:          phone,
		AccountTypeID:  accountTypeID,
		EmailVerified:  false,
		PhoneVerified:  false,
		IdentityVerified: false,
		KYCStatus:      "pending",
		IsActive:       true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

// ============================================================
// VERIFICATION METHODS
// ============================================================

// VerifyEmail marks the user's email as verified
func (u *User) VerifyEmail() {
	now := time.Now()
	u.EmailVerified = true
	u.EmailVerifiedAt = &now
	u.UpdatedAt = now
}

// VerifyPhone marks the user's phone as verified
func (u *User) VerifyPhone() {
	now := time.Now()
	u.PhoneVerified = true
	u.PhoneVerifiedAt = &now
	u.UpdatedAt = now
}

// VerifyIdentity marks the user's identity as verified
func (u *User) VerifyIdentity() {
	now := time.Now()
	u.IdentityVerified = true
	u.IdentityVerifiedAt = &now
	u.UpdatedAt = now
}

// IsVerified returns true if email, phone, and identity are all verified
func (u *User) IsVerified() bool {
	return u.EmailVerified && u.PhoneVerified && u.IdentityVerified
}

// IsEmailVerified returns true if email is verified
func (u *User) IsEmailVerified() bool {
	return u.EmailVerified
}

// ============================================================
// KYC METHODS
// ============================================================

// SubmitKYC submits KYC documents
func (u *User) SubmitKYC(idDocument, selfieDocument, addressProof string) {
	now := time.Now()
	u.KYCStatus = "submitted"
	u.KYCSubmittedAt = &now
	u.IDDocument = &idDocument
	u.SelfieDocument = &selfieDocument
	u.AddressProof = &addressProof
	u.UpdatedAt = now
}

// VerifyKYC marks KYC as verified
func (u *User) VerifyKYC() {
	now := time.Now()
	u.KYCStatus = "verified"
	u.KYCVerifiedAt = &now
	u.IdentityVerified = true
	u.IdentityVerifiedAt = &now
	u.UpdatedAt = now
}

// RejectKYC rejects KYC with a reason
func (u *User) RejectKYC(reason string) {
	now := time.Now()
	u.KYCStatus = "rejected"
	u.KYCRejectedAt = &now
	u.KYCRejectionReason = &reason
	u.UpdatedAt = now
}

// IsKYCVerified checks if KYC is verified
func (u *User) IsKYCVerified() bool {
	return u.KYCStatus == "verified"
}

// IsKYCRequired checks if KYC is required
func (u *User) IsKYCRequired() bool {
	return u.KYCStatus != "not_required"
}

// ============================================================
// STATUS METHODS
// ============================================================

// Activate activates the user
func (u *User) Activate() error {
	if u.IsActive {
		return errors.New("user already active")
	}
	if !u.EmailVerified {
		return errors.New("cannot activate unverified user")
	}
	u.IsActive = true
	u.UpdatedAt = time.Now()
	return nil
}

// Deactivate deactivates the user
func (u *User) Deactivate() error {
	if !u.IsActive {
		return errors.New("user already inactive")
	}
	u.IsActive = false
	u.UpdatedAt = time.Now()
	return nil
}

// IsActiveUser returns true if the user is active
func (u *User) IsActiveUser() bool {
	return u.IsActive
}

// ============================================================
// PROFILE UPDATE METHODS
// ============================================================

// UpdateName updates the user's name
func (u *User) UpdateName(name string) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}
	u.Name = name
	u.UpdatedAt = time.Now()
	return nil
}

// UpdateDisplayName updates the user's display name
func (u *User) UpdateDisplayName(displayName string) error {
	if displayName == "" {
		return errors.New("display name cannot be empty")
	}
	u.DisplayName = displayName
	u.UpdatedAt = time.Now()
	return nil
}

// UpdateEmail updates the user's email
func (u *User) UpdateEmail(email string) error {
	if email == "" {
		return errors.New("email cannot be empty")
	}
	u.Email = email
	u.EmailVerified = false
	u.EmailVerifiedAt = nil
	u.UpdatedAt = time.Now()
	return nil
}

// UpdatePhone updates the user's phone
func (u *User) UpdatePhone(phone string) error {
	if phone == "" {
		return errors.New("phone cannot be empty")
	}
	u.Phone = phone
	u.PhoneVerified = false
	u.PhoneVerifiedAt = nil
	u.UpdatedAt = time.Now()
	return nil
}

// UpdatePassword updates the user's password
func (u *User) UpdatePassword(hashedPassword string) {
	u.PasswordHash = hashedPassword
	u.UpdatedAt = time.Now()
}

// UpdateProfile updates multiple profile fields at once
func (u *User) UpdateProfile(name, displayName, phone string) {
	if name != "" {
		u.Name = name
	}
	if displayName != "" {
		u.DisplayName = displayName
	}
	if phone != "" {
		u.Phone = phone
	}
	u.UpdatedAt = time.Now()
}

// ============================================================
// INSTITUTION METHODS
// ============================================================

// UpdateInstitution updates the user's institution
func (u *User) UpdateInstitution(institutionID *string) {
	u.InstitutionID = institutionID
	u.UpdatedAt = time.Now()
}

// SetInstitution sets the user's institution
func (u *User) SetInstitution(institutionID string) {
	u.InstitutionID = &institutionID
	u.UpdatedAt = time.Now()
}

// RemoveInstitution removes the user's institution
func (u *User) RemoveInstitution() {
	u.InstitutionID = nil
	u.UpdatedAt = time.Now()
}

// ============================================================
// SOFT DELETE METHODS
// ============================================================

// SoftDelete soft deletes the user
func (u *User) SoftDelete() {
	now := time.Now()
	u.DeletedAt = &now
	u.IsActive = false
	u.UpdatedAt = now
}

// Restore restores a soft-deleted user
func (u *User) Restore() {
	u.DeletedAt = nil
	u.IsActive = true
	u.UpdatedAt = time.Now()
}

// ============================================================
// HELPER FUNCTIONS
// ============================================================

// slugify generates a slug from a name
func slugify(name string) string {
	// Simple slug generation - can be improved
	// In production, you'd want to check for uniqueness
	return "user-" + uuid.New().String()[:8]
}