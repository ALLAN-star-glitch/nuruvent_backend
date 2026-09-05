// internal/modules/account/service/role_manager.go

package service

import "context"

// RoleManager defines the role management operations needed by the account module
type RoleManager interface {
    // AssignRole assigns a role to a user in a scope
    AssignRole(ctx context.Context, scope string, userID, role string) error

    // RemoveRole removes a role from a user in a scope
    RemoveRole(ctx context.Context, scope string, userID, role string) error

    // GetUserRoles returns all roles for a user in a scope
    GetUserRoles(ctx context.Context, scope string, userID string) ([]string, error)

    // HasRole checks if a user has a specific role in a scope
    HasRole(ctx context.Context, scope string, userID, role string) (bool, error)
}

// Scope represents a permission scope (account scope)
type Scope interface {
    IsPersonal() bool
    IsInstitution() bool
    GetID() string
    String() string
}

// AccountScope represents an account scope
type AccountScope struct {
    AccountID string
}

func NewAccountScope(accountID string) AccountScope {
    return AccountScope{AccountID: accountID}
}

func (s AccountScope) IsPersonal() bool {
    return false
}

func (s AccountScope) IsInstitution() bool {
    return false
}

func (s AccountScope) GetID() string {
    return s.AccountID
}

func (s AccountScope) String() string {
    return "account:" + s.AccountID
}