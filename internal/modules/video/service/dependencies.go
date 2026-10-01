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

	// Cross-module
	//
	// Attendance is the video module's view of the attendance
	// service. It is used only by FetchGoogleMeetAttendance, which
	// hands off fetched participants for the attendance module to
	// record.
	//
	// The interface is defined in attendance_contract.go, in this
	// same package.
	Attendance ParticipantRecorder

	// Cross-cutting
	IDs   *id.UUIDGenerator
	Clock Clock
}