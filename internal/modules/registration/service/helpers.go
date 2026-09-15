package service

import "time"

// SystemClock is the production Clock.
type SystemClock struct{}

func (SystemClock) Now() time.Time { return time.Now() }

// The IDGenerator implementation will live in infrastructure (it needs uuid).
// We only define the interface here (already done in service.go).