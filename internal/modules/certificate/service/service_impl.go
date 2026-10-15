// internal/modules/certificate/service/service_impl.go

package service

// serviceImpl is the concrete implementation of Service.
//
// Unexported: callers depend on the Service interface. NewService
// returns the interface so wire binds it without a wire.Bind.
type serviceImpl struct {
	deps Dependencies
}

// NewService constructs the certificate Service.
func NewService(deps Dependencies) Service {
	return &serviceImpl{deps: deps}
}

// compile-time assertion
var _ Service = (*serviceImpl)(nil)