// internal/modules/certificate/domain/id_generator.go

package domain

// IDGenerator produces identifiers for new aggregates.
//
// The domain defines this port itself so it stays free of third-party
// dependencies. Adapters wrap internal/shared/id.Generator to satisfy it.
type IDGenerator interface {
	NewID() string
}