// internal/modules/account/accountdomain/context_keys.go

package accountdomain

import "github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/types"

// ============================================================
// RE-EXPORTS from internal/shared/types
// ============================================================
//
// The account module needs access to the same context keys and helpers
// as auth and events. Rather than duplicating them, re-export from the
// shared package.

type ContextKey = types.ContextKey

const (
	ContextKeyUserID      = types.ContextKeyUserID
	ContextKeyUserRole    = types.ContextKeyUserRole
	ContextKeyUserEmail   = types.ContextKeyUserEmail
	ContextKeyUserName    = types.ContextKeyUserName
	ContextKeyDomain      = types.ContextKeyDomain
	ContextKeyAccountID   = types.ContextKeyAccountID
	ContextKeyAccountType = types.ContextKeyAccountType
	ContextKeyTeamID      = types.ContextKeyTeamID
	ContextKeyTeamType    = types.ContextKeyTeamType
)

var (
	WithUserID         = types.WithUserID
	GetUserID          = types.GetUserID
	WithTeamContext    = types.WithTeamContext
	GetTeamID          = types.GetTeamID
	GetTeamType        = types.GetTeamType
	WithAccountContext = types.WithAccountContext
	GetAccountID       = types.GetAccountID
	GetAccountType     = types.GetAccountType
	WithDomain         = types.WithDomain
	GetDomain          = types.GetDomain
)