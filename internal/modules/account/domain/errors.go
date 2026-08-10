package domain

import "errors"

var (
    ErrAccountNotFound = errors.New("account not found")
    ErrAccountTypeNotFound = errors.New("account type not found")   
    ErrAccountExists   = errors.New("account already exists")
    ErrInvalidEmail    = errors.New("invalid email address")
    ErrInvalidPassword = errors.New("invalid password")
    ErrInvalidPhone    = errors.New("invalid phone number")
    ErrAccountInactive = errors.New("account is inactive")
    ErrInvalidAccountType = errors.New("invalid account type")
    ErrInstitutionTypeNotFound = errors.New("institution type not found")
    ErrInvalidProfessionalType = errors.New("invalid professional type")
    ErrProfessionalTypeNotFound = errors.New("professional type not found")
    ErrInvalidInstitutionType = errors.New("invalid institution type")
    ErrTeamMemberNotFound = errors.New("team member not found") 
    ErrInstitutionNotFound = errors.New("institution not found")
    ErrTeamMemberExists = errors.New("team member already exists")
    ErrInvalidRole = errors.New("invalid role")

)