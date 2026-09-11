// internal/modules/account/accountdomain/repository.go

package accountdomain

import "context"

// Repository defines the data access interface for the account module
type Repository interface {
    // Account Type operations
    GetAccountTypeByID(ctx context.Context, id string) (*AccountType, error)
    GetAccountTypeByName(ctx context.Context, name string) (*AccountType, error)
    GetAccountTypeBySlug(ctx context.Context, slug string) (*AccountType, error)
    GetAccountTypes(ctx context.Context) ([]*AccountType, error)
    GetAccountIDByTeamID(ctx context.Context, teamID string) (string, error)

    // Account operations
    CreateAccount(ctx context.Context, account *Account) error
    GetAccountByID(ctx context.Context, id string) (*Account, error)
    GetAccountByEmail(ctx context.Context, email string) (*Account, error)
    GetAccountBySlug(ctx context.Context, slug string) (*Account, error)
    GetAccountsByUserID(ctx context.Context, userID string) ([]*Account, error)
    UpdateAccount(ctx context.Context, account *Account) error
    DeleteAccount(ctx context.Context, id string) error

    // Account Member operations
    CreateAccountMember(ctx context.Context, member *AccountMember) error
    GetAccountMemberByID(ctx context.Context, id string) (*AccountMember, error)
    GetAccountMemberByAccountAndUser(ctx context.Context, accountID, userID string) (*AccountMember, error)
    GetAccountMembersByAccount(ctx context.Context, accountID string) ([]*AccountMember, error)
    GetAccountMembersByUser(ctx context.Context, userID string) ([]*AccountMember, error)
    UpdateAccountMember(ctx context.Context, member *AccountMember) error
    DeleteAccountMember(ctx context.Context, accountID, userID string) error
    CountAccountMembers(ctx context.Context, accountID string) (int64, error)
}