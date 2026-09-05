// internal/modules/account/service/service.go

package service

import (
    "context"

    "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/accountdomain"
)

// Service defines the account module's business logic interface
type Service interface {
    // Account Type operations
    GetAccountTypes(ctx context.Context) ([]*accountdomain.AccountType, error)
    GetAccountTypeByID(ctx context.Context, id string) (*accountdomain.AccountType, error)
    GetAccountTypeBySlug(ctx context.Context, slug string) (*accountdomain.AccountType, error)

    // Account operations
    CreatePersonalAccount(ctx context.Context, cmd CreatePersonalAccountCommand) (*accountdomain.Account, error)
    CreateInstitutionAccount(ctx context.Context, cmd CreateInstitutionAccountCommand) (*accountdomain.Account, error)
    GetAccountByID(ctx context.Context, id string) (*accountdomain.Account, error)
    GetAccountBySlug(ctx context.Context, slug string) (*accountdomain.Account, error)
    GetUserAccounts(ctx context.Context, userID string) ([]*accountdomain.Account, error)
    UpdateAccount(ctx context.Context, id string, updates map[string]interface{}) (*accountdomain.Account, error)
    DeleteAccount(ctx context.Context, id string) error

    // Account Member operations
    AddMember(ctx context.Context, cmd AddMemberCommand) (*accountdomain.AccountMember, error)
    RemoveMember(ctx context.Context, accountID, userID, removedBy string) error
    UpdateMemberRole(ctx context.Context, accountID, userID, newRole, updatedBy string) (*accountdomain.AccountMember, error)
    GetAccountMembers(ctx context.Context, accountID string) ([]*accountdomain.AccountMember, error)
    LeaveAccount(ctx context.Context, accountID, userID string) error
}

// ============================================================
// COMMANDS
// ============================================================

// CreatePersonalAccountCommand - used when a user creates a personal account
type CreatePersonalAccountCommand struct {
    Name      string // User's name (e.g., "John Doe")
    Email     string // User's email
    Phone     string // User's phone
    CreatedBy string // User ID of the creator
}

// CreateInstitutionAccountCommand - used when a user creates an institution account
type CreateInstitutionAccountCommand struct {
    Name        string // Institution name (e.g., "Nuruvent Ltd")
    Email       string // Institution email (must be professional)
    Phone       string // Institution phone
    Website     string // Institution website
    Description string // Institution description
    CreatedBy   string // User ID of the creator
}

// AddMemberCommand - used when adding a member to an account
type AddMemberCommand struct {
    AccountID string
    UserID    string
    Role      string // "account_admin" or "trainer"
    InvitedBy string
}

// UpdateAccountCommand - used when updating an account
type UpdateAccountCommand struct {
    Name        string
    Email       string
    Phone       string
    Website     string
    Description string
    LogoURL     string
}