//go:build wireinject
// +build wireinject

package account

import (
	"github.com/google/wire"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/delivery/acchandler"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/postgres"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/service"
)

var ProviderSet = wire.NewSet(
	postgres.NewPostgresRepository,
	service.NewService,
	acchandler.NewAccountHandler,
)