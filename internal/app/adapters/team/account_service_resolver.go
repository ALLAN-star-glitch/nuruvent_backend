package team

import (
	"errors"
	"sync"

	accountService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/service"
)

// AccountServiceResolver is the func type that yields the account
// service on demand.
type AccountServiceResolver func() (accountService.Service, error)

// AccountServiceHolder stores a reference to the account service so
// team adapters can reach it after wire finishes.
//
// The account service depends on the auth service, which depends on
// the team service, which depends on the team's UserProfileProvider
// adapter — a cycle Wire cannot resolve. This holder breaks it by
// deferring the account service lookup until first use.
//
// Wire provides a *AccountServiceHolder, then app.go calls
// holder.Set(svc) once the graph is built. Adapters read through
// holder.Get().
type AccountServiceHolder struct {
	mu  sync.RWMutex
	svc accountService.Service
}

// NewAccountServiceHolder returns an empty holder.
func NewAccountServiceHolder() *AccountServiceHolder {
	return &AccountServiceHolder{}
}

// Set stores the account service. Safe to call once, after wire.
func (h *AccountServiceHolder) Set(svc accountService.Service) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.svc = svc
}

// Get returns the account service or an error if it hasn't been set.
func (h *AccountServiceHolder) Get() (accountService.Service, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.svc == nil {
		return nil, errors.New("account service not yet initialized")
	}
	return h.svc, nil
}

// Resolver returns a func that satisfies AccountServiceResolver.
func (h *AccountServiceHolder) Resolver() AccountServiceResolver {
	return func() (accountService.Service, error) {
		return h.Get()
	}
}