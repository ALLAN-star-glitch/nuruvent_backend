// internal/shared/config/providers.go

package config

import "github.com/google/wire"

// ProviderSet contains all config module dependencies
var ProviderSet = wire.NewSet(
	Load,
)