// internal/modules/media/providers.go

package media

import (
	"github.com/google/wire"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/media/postgres"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/media/service"
)

var ProviderSet = wire.NewSet(

	// Repository
	postgres.NewPostgresRepository,

	// Service
	service.NewService,
)