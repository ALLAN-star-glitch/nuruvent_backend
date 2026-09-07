// internal/modules/team/provider.go

package team

import (
	"github.com/google/wire"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/team/delivery/handler"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/team/infrastructure/postgres"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/team/service"
)

// ============================================================
// PROVIDER SET
// ============================================================

var ProviderSet = wire.NewSet(
	// Repository
	postgres.NewTeamRepository,

	// Service
	service.NewTeamService,

	// Handler
	handler.NewTeamHandler,
)

// ============================================================
// INTERFACE BINDINGS (if needed)
// ============================================================

// Note: Wire will automatically match types:
// - postgres.NewTeamRepository returns *postgres.TeamRepository which implements teamdomain.Repository
// - service.NewTeamService returns service.Service interface
// - handler.NewTeamHandler returns *handler.TeamHandler