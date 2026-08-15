// internal/shared/queue/providers.go

package queue

import "github.com/google/wire"

// ProviderSet contains all queue module dependencies
var ProviderSet = wire.NewSet(
	NewClient, // ✅ Returns interface, no wire.Bind needed
)