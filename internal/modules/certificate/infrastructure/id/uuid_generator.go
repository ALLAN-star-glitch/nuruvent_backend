// internal/modules/certificate/infrastructure/id/uuid_generator.go

package id

import (
	sharedid "github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/id"
)

// UUIDGenerator adapts internal/shared/id.Generator to the certificate
// module's domain.IDGenerator port.
//
// The domain defines its own IDGenerator interface so it stays free of
// third-party dependencies; this adapter bridges the two.
type UUIDGenerator struct {
	inner sharedid.Generator
}

// NewUUIDGenerator constructs a UUIDGenerator backed by the shared
// ID generator.
func NewUUIDGenerator(inner sharedid.Generator) *UUIDGenerator {
	return &UUIDGenerator{inner: inner}
}

// NewID implements domain.IDGenerator.
func (g *UUIDGenerator) NewID() string {
	return g.inner.NewID()
}