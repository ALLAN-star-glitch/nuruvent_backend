// internal/modules/events/domain/context_keys.go

package domain

import "github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/types"

// ============================================================
// RE-EXPORTS from internal/shared/types
// ============================================================
//
// The canonical implementation lives in internal/shared/types. This file
// re-exports it under the events domain namespace so existing callsites
// continue to work unchanged.

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