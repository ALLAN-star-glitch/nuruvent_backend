// internal/modules/certificate/infrastructure/clock/system_clock.go

package clock

import "time"

// SystemClock adapts time.Now to the certificate module's domain.Clock
// port.
type SystemClock struct{}

// NewSystemClock constructs a SystemClock.
func NewSystemClock() *SystemClock { return &SystemClock{} }

// Now implements domain.Clock.
func (*SystemClock) Now() time.Time {
	return time.Now().UTC()
}