// internal/modules/certificate/domain/clock.go

package domain

import "time"

// Clock abstracts time so the application service is deterministic
// under test. Adapters provide a real implementation backed by time.Now.
type Clock interface {
	Now() time.Time
}