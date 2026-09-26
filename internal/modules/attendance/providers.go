// internal/modules/attendance/providers.go

package attendance

import (
	"time"

	"github.com/google/wire"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
	attendanceHandler "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/delivery/http"
	attendancePostgres "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/infrastructure/postgres"
	attendanceZoom "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/infrastructure/providers/zoom"
	attendancePublisher "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/infrastructure/publisher"
	attendanceToken "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/infrastructure/token"
	attendanceService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/service"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/config"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/id"
)

// ProviderSet is the attendance module's wire provider set.
var ProviderSet = wire.NewSet(
	// Persistence
	attendancePostgres.NewAttendeeRepository,
	attendancePostgres.NewSessionRepository,
	attendancePostgres.NewJoinTokenRepository,
	attendancePostgres.NewAttendanceRecordRepository,
	attendancePostgres.NewAttendeeSessionStatusRepository,
	attendancePostgres.NewAttendeeRollupStatusRepository,
	attendancePostgres.NewAttendanceOverrideRepository,
	attendancePostgres.NewUnitOfWork,

	// Cross-cutting adapters
	attendanceToken.NewSHA256Generator,
	attendancePublisher.NewLogPublisher,
	NewSystemClock,
	wire.Bind(new(attendanceService.Clock), new(*systemClock)),

	// Video providers
	ProvideAttendanceProviders,

	// Service dependencies
	ProvideAttendanceDependencies,

	// Service
	attendanceService.NewService,

	// HTTP handlers
	attendanceHandler.NewSessionHandler,
	attendanceHandler.NewAttendanceHandler,
	attendanceHandler.NewWebhookHandler,
	attendanceHandler.NewJoinHandler,
	attendanceHandler.NewHandlers,
)

// ============================================================
// CLOCK
// ============================================================

// systemClock implements attendanceService.Clock by delegating to
// time.Now.
type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now().UTC() }

// newSystemClock constructs the concrete clock. The ProviderSet binds
// *systemClock to the attendanceService.Clock interface.
func NewSystemClock() *systemClock {
	return &systemClock{}
}

// ============================================================
// VIDEO PROVIDERS
// ============================================================

// provideAttendanceProviders builds the map of video-platform
// adapters.
func ProvideAttendanceProviders(
	cfg *config.Config,
) map[attendance.SessionProvider]attendanceService.ProviderAdapter {
	providers := make(map[attendance.SessionProvider]attendanceService.ProviderAdapter)

	if cfg.Zoom.IsConfigured() {
		providers[attendance.ProviderZoom] = attendanceZoom.NewProvider(cfg.Zoom)
	}

	return providers
}

// ============================================================
// SERVICE DEPENDENCIES
// ============================================================

// provideAttendanceDependencies assembles service.Dependencies from
// the wire-bound building blocks.
func ProvideAttendanceDependencies(
	unitOfWork *attendancePostgres.UnitOfWork,
	ids *id.UUIDGenerator,
	tokens *attendanceToken.SHA256Generator,
	publisher *attendancePublisher.LogPublisher,
	clock *systemClock,
	providers map[attendance.SessionProvider]attendanceService.ProviderAdapter,
) attendanceService.Dependencies {
	return attendanceService.Dependencies{
		UnitOfWork:       unitOfWork,
		Clock:            clock,
		IDs:              ids,
		TokenGenerator:   tokens,
		Publisher:        publisher,
		Providers:        providers,
		DerivationPolicy: attendance.DefaultDerivationPolicy(),
		JoinTokenGrace:   24 * time.Hour,
	}
}