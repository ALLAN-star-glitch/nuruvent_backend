// internal/shared/redis/providers.go

package redis

import "github.com/google/wire"

// ProviderSet contains all redis module dependencies
var ProviderSet = wire.NewSet(
	NewClient,
)