// internal/shared/id/providers.go

package id

import "github.com/google/wire"

// ProviderSet exposes the ID generator for the whole app.
//
// NewUUIDGenerator returns *UUIDGenerator. The wire.Bind tells Wire
// that *UUIDGenerator satisfies the Generator interface, so consumers
// that depend on id.Generator can be resolved.
var ProviderSet = wire.NewSet(
	NewUUIDGenerator,
	wire.Bind(new(Generator), new(*UUIDGenerator)),
)
