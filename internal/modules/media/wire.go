//go:build wireinject
// +build wireinject

package media

import (
	"github.com/google/wire"


	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/media/postgres"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/media/service"

)

var ProviderSet = wire.NewSet(
	postgres.NewPostgresRepository,
	service.NewService,
)
