package media

import (
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/config"
	"github.com/ALLAN_star_glitch/nuruvent-backend/pkg/storage"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	ProvideMediaModule,
)

func ProvideMediaModule(
	cfg *config.Config,
	storageClient *storage.Client,
) *Module {
	return NewModule(cfg, storageClient)
}