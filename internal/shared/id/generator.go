package id

import "github.com/google/uuid"

// Generator produces new IDs. Any service that needs unique IDs depends on
// this interface. The concrete UUIDGenerator below is the default impl.
type Generator interface {
	NewID() string
}

// UUIDGenerator produces random UUID v4 strings.
type UUIDGenerator struct{}

func NewUUIDGenerator() *UUIDGenerator { return &UUIDGenerator{} }

func (*UUIDGenerator) NewID() string { return uuid.NewString() }