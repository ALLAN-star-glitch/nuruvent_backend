// internal/modules/video/providers.go

package video

import (
	"fmt"
	"time"

	"github.com/google/wire"
	"gorm.io/gorm"

	videoHandler "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/delivery/http"
	videoCrypto "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/infrastructure/crypto"
	videoPostgres "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/infrastructure/postgres"
	googleMeetProvider "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/infrastructure/providers/googlemeet"
	zoomProvider "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/infrastructure/providers/zoom"
	videodomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/videodomain"
	videoService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/service"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/config"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/id"
)

// ProviderSet is the video module's wire provider set.
var ProviderSet = wire.NewSet(
	// Persistence
	videoPostgres.NewConnectionRepository,
	wire.Bind(new(videodomain.ConnectionRepository), new(*videoPostgres.ConnectionRepository)),

	videoPostgres.NewOAuthStateRepository,
	wire.Bind(new(videodomain.OAuthStateRepository), new(*videoPostgres.OAuthStateRepository)),

	videoPostgres.NewMeetingRepository,
	wire.Bind(new(videodomain.MeetingRepository), new(*videoPostgres.MeetingRepository)),

	videoPostgres.NewUnitOfWork,
	wire.Bind(new(videodomain.UnitOfWork), new(*videoPostgres.UnitOfWork)),

	// Crypto
	ProvideTokenCipher,

	// Clock
	NewSystemClock,
	wire.Bind(new(videoService.Clock), new(*systemClock)),

	// Provider clients
	ProvideZoomClient,
	ProvideGoogleMeetClient,
	ProvideClientRegistry,

	// Service dependencies
	ProvideVideoDependencies,

	// Service
	videoService.New,

	// HTTP handlers
	videoHandler.NewHandlers,
)

// ============================================================
// CLOCK
// ============================================================

// systemClock implements videoService.Clock by delegating to
// time.Now.
type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now().UTC() }

// NewSystemClock constructs the concrete clock. The ProviderSet binds
// *systemClock to the videoService.Clock interface.
func NewSystemClock() *systemClock {
	return &systemClock{}
}

// ============================================================
// CRYPTO
// ============================================================

func ProvideTokenCipher(cfg *config.Config) (videodomain.TokenCipher, error) {
	if cfg.Video.EncryptionKey == "" {
		return nil, fmt.Errorf("video: VIDEO_TOKEN_ENCRYPTION_KEY is required")
	}
	return videoCrypto.NewAESGCMCipher(cfg.Video.EncryptionKey)
}

// ============================================================
// PROVIDER CLIENT TYPES
// ============================================================
//
// Wire resolves dependencies by type identity. Two parameters of the
// same interface type (videodomain.ProviderClient) are ambiguous —
// wire cannot decide which provider's output goes to which parameter.
//
// These named types give each platform's client a distinct identity
// in the injection graph. They are aliases in behavior: any value of
// ZoomClient or GoogleMeetClient satisfies videodomain.ProviderClient,
// because the underlying type is the same interface. The distinction
// exists only for wire's benefit.
type (
	// ZoomClient is the Zoom provider client, distinct in wire's
	// graph from GoogleMeetClient.
	ZoomClient videodomain.ProviderClient

	// GoogleMeetClient is the Google Meet provider client, distinct
	// in wire's graph from ZoomClient.
	GoogleMeetClient videodomain.ProviderClient
)

// ============================================================
// PROVIDER CLIENTS
// ============================================================

// ProvideZoomClient constructs the Zoom client from config, or
// returns nil if Zoom OAuth isn't configured.
//
// The return type is the named ZoomClient so wire can distinguish it
// from GoogleMeetClient.
func ProvideZoomClient(cfg *config.Config) ZoomClient {
	if !cfg.Video.ZoomOAuth.IsConfigured() {
		return nil
	}
	return zoomProvider.NewClient(cfg.Video.ZoomOAuth)
}

// ProvideGoogleMeetClient constructs the Google Meet client from
// config, or returns nil if Google Meet OAuth isn't configured.
//
// The return type is the named GoogleMeetClient so wire can
// distinguish it from ZoomClient.
func ProvideGoogleMeetClient(cfg *config.Config) GoogleMeetClient {
	if !cfg.Video.GoogleMeetOAuth.IsConfigured() {
		return nil
	}
	return googleMeetProvider.NewClient(cfg.Video.GoogleMeetOAuth)
}

// ProvideClientRegistry builds the registry from configured clients.
//
// Platforms that aren't configured are simply not registered. The
// service returns ErrUnsupportedPlatform when asked for one.
//
// The parameters are the named types, so wire can resolve them
// unambiguously. Inside, they are converted back to the interface
// type for storage in the registry map.
func ProvideClientRegistry(
	zoom ZoomClient,
	googleMeet GoogleMeetClient,
) videodomain.ClientRegistry {
	clients := make(map[videodomain.Platform]videodomain.ProviderClient)

	if zoom != nil {
		clients[videodomain.PlatformZoom] = videodomain.ProviderClient(zoom)
	}
	if googleMeet != nil {
		clients[videodomain.PlatformGoogleMeet] = videodomain.ProviderClient(googleMeet)
	}

	return videoService.NewClientRegistry(clients)
}

// ============================================================
// SERVICE DEPENDENCIES
// ============================================================

func ProvideVideoDependencies(
	unitOfWork *videoPostgres.UnitOfWork,
	connections videodomain.ConnectionRepository,
	oauthStates videodomain.OAuthStateRepository,
	meetings videodomain.MeetingRepository,
	clients videodomain.ClientRegistry,
	ids *id.UUIDGenerator,
	clock *systemClock,
) videoService.Dependencies {
	return videoService.Dependencies{
		UnitOfWork:  unitOfWork,
		Connections: connections,
		OAuthStates: oauthStates,
		Meetings:    meetings,
		Clients:     clients,
		IDs:         ids,
		Clock:       clock,
	}
}

// ============================================================
// MISC
// ============================================================

// ensure gorm is imported; the module's repositories use it.
var _ = gorm.ErrRecordNotFound