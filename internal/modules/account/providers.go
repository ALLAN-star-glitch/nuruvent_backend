// internal/modules/account/provider.go

package account

import (
	"github.com/google/wire"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/infrastructure/postgres"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/delivery/handler"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/service"
)

// ============================================================
// PROVIDER SET
// ============================================================

var ProviderSet = wire.NewSet(
	// Repository
	postgres.NewAccountRepository,

	// Service
	service.NewAccountService,

	// Handler
	handler.NewAccountHandler,
	
)