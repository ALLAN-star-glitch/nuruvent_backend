// internal/shared/database/providers.go

package database

import "github.com/google/wire"

// ProviderSet contains all database module dependencies
var ProviderSet = wire.NewSet(
	Connect,
)