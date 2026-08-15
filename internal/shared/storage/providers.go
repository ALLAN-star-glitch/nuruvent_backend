// internal/shared/storage/providers.go

package storage

import "github.com/google/wire"

// ProviderSet contains all storage module dependencies
var ProviderSet = wire.NewSet(
	NewClientFromConfig,
)