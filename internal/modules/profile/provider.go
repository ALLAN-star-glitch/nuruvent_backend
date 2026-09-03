// internal/modules/profile/providers.go

package profile

import (
	"github.com/google/wire"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/profile/delivery/handler"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/profile/infrastructure/postgres"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/profile/service"
)

var ProviderSet = wire.NewSet(
	// Repository
	postgres.NewProfileRepository,
	
	// Service
	service.NewProfileService,

	handler.NewProfileHandler,
)