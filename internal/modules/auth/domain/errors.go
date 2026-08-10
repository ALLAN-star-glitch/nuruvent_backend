package domain

import "errors"

// ============================================================
// AUTH DOMAIN ERRORS
// ============================================================

// Account errors
var (
    ErrAccountNotFound      = errors.New("account not found")
    ErrAccountExists        = errors.New("account already exists")
    ErrAccountInactive      = errors.New("account is inactive")
    ErrAccountAlreadyActive = errors.New("account already active")
    ErrAccountNotVerified   = errors.New("account email not verified")
)

// Credential errors
var (
    ErrInvalidEmail    = errors.New("invalid email address")
    ErrInvalidPassword = errors.New("invalid password")
    ErrInvalidPhone    = errors.New("invalid phone number")
    ErrWeakPassword    = errors.New("password too weak")
)

// Authentication errors
var (
    ErrInvalidCredentials  = errors.New("invalid credentials")
    ErrAccountLocked       = errors.New("account is locked")
    ErrTooManyAttempts     = errors.New("too many failed attempts")
)

// OTP errors
var (
    ErrInvalidOTP     = errors.New("invalid OTP")
    ErrExpiredOTP     = errors.New("OTP has expired")
    ErrOTPNotFound    = errors.New("OTP not found")
)

// Token errors
var (
    ErrInvalidToken   = errors.New("invalid token")
    ErrExpiredToken   = errors.New("token has expired")
    ErrRevokedToken   = errors.New("token has been revoked")
    ErrTokenNotFound  = errors.New("token not found")
)

// Account type errors
var (
    ErrInvalidAccountType     = errors.New("invalid account type")
    ErrAccountTypeNotFound    = errors.New("account type not found")
    ErrProfessionalTypeNotFound = errors.New("professional type not found")
)

// Institution errors
var (
    ErrInstitutionNotFound = errors.New("institution not found")
    ErrInstitutionExists   = errors.New("institution already exists")
)

// Team member errors
var (
    ErrTeamMemberNotFound = errors.New("team member not found")
    ErrTeamMemberExists   = errors.New("team member already exists")
    ErrInvalidRole        = errors.New("invalid role")
)