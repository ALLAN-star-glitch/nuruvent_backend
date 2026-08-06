// internal/modules/media/module.go

package media

import (
	"context"
	"log"

	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/config"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/database"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/media/mediarepo"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/media/mediaservice"
	"github.com/ALLAN_star_glitch/nuruvent-backend/pkg/storage"
)

type Module struct {
	service *mediaservice.MediaService
}

func NewModule(
	cfg *config.Config,
	storageClient *storage.Client,
) *Module {
	db := database.GetDB()

	// Initialize repositories
	mediaRepo := mediarepo.NewMediaRepository(db)
	mediaTypeRepo := mediarepo.NewMediaTypeRepository(db)

	// Initialize service (no enforcer needed - pure file operations)
	mediaService := mediaservice.NewMediaService(
		mediaRepo,
		mediaTypeRepo,
		storageClient,
	)

	return &Module{
		service: mediaService,
	}
}

// GetService returns the media service
func (m *Module) GetService() *mediaservice.MediaService {
	return m.service
}

func (m *Module) Init(ctx context.Context) error {
	log.Println("Media module initialized")
	return nil
}

func (m *Module) Close() {
	log.Println("Media module closed")
}