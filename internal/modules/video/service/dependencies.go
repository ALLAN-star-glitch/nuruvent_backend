// internal/modules/video/service/dependencies.go

package service

import (
	"time"

	videodomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/videodomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/id"
)

// ============================================================
// CLOCK
// ============================================================

// Clock abstracts time.Now so tests can control time.
type Clock interface {
	Now() time.Time
}

// ============================================================
// DEPENDENCIES
// ============================================================

// Dependencies is the full set of ports the service needs.
//
// Wired together by providers.go; the concrete service constructor
// accepts this struct and stores it.
type Dependencies struct {
	// Persistence
	UnitOfWork  videodomain.UnitOfWork
	Connections videodomain.ConnectionRepository
	OAuthStates videodomain.OAuthStateRepository
	Meetings    videodomain.MeetingRepository

	// Platform clients
	Clients videodomain.ClientRegistry

	// Cross-cutting
	IDs   *id.UUIDGenerator
	Clock Clock
}