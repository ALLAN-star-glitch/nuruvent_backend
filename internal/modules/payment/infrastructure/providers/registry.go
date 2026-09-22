
// internal/modules/payment/infrastructure/providers/registry.go

package providers

import (
	"fmt"
	"sync"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
)

// Registry is the default ProviderRegistry implementation.
//
// It holds a set of registered providers, indexed by name and by
// method. Providers register at construction time; the registry is
// read-only afterward, so it's safe for concurrent use.
type Registry struct {
	mu        sync.RWMutex
	byName    map[string]paymentdomain.PaymentProvider
	byMethod  map[paymentdomain.PaymentMethod]paymentdomain.PaymentProvider
}

// NewRegistry constructs an empty registry.
func NewRegistry() *Registry {
	return &Registry{
		byName:   make(map[string]paymentdomain.PaymentProvider),
		byMethod: make(map[paymentdomain.PaymentMethod]paymentdomain.PaymentProvider),
	}
}

// Register adds a provider to the registry. The first provider
// registered for a given method becomes the default for ByMethod.
//
// Returns an error if a provider with the same name is already
// registered. Overwriting is not allowed — the registry is built once
// at composition time.
func (r *Registry) Register(p paymentdomain.PaymentProvider) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := p.Name()
	if name == "" {
		return fmt.Errorf("provider name must not be empty")
	}
	if _, exists := r.byName[name]; exists {
		return fmt.Errorf("provider %q already registered", name)
	}

	r.byName[name] = p

	method := p.Method()
	if _, exists := r.byMethod[method]; !exists {
		r.byMethod[method] = p
	}

	return nil
}

// ByName returns the provider registered under the given name.
func (r *Registry) ByName(name string) (paymentdomain.PaymentProvider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.byName[name]
	if !ok {
		return nil, fmt.Errorf("%w: %q", paymentdomain.ErrProviderNotFound, name)
	}
	return p, nil
}

// ByMethod returns the default provider that serves the given method.
func (r *Registry) ByMethod(method paymentdomain.PaymentMethod) (paymentdomain.PaymentProvider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.byMethod[method]
	if !ok {
		return nil, fmt.Errorf("%w: method %q", paymentdomain.ErrProviderNotFound, method)
	}
	return p, nil
}

// List returns all registered providers.
func (r *Registry) List() []paymentdomain.PaymentProvider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]paymentdomain.PaymentProvider, 0, len(r.byName))
	for _, p := range r.byName {
		out = append(out, p)
	}
	return out
}

// Compile-time assertion.
var _ paymentdomain.ProviderRegistry = (*Registry)(nil)