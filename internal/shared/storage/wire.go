//go:build wireinject
// +build wireinject

package storage

import (
	"github.com/google/wire"
	"gorm.io/gorm"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/config"
)

// ============================================================
// PROVIDER SET
// ============================================================

var ProviderSet = wire.NewSet(
	ProvideMediaRepository,
	ProvideMediaTypeRepository,
	ProvideStorageClient,
	ProvideMediaService,
)

// ============================================================
// PROVIDERS
// ============================================================

func ProvideMediaRepository(db *gorm.DB) *MediaRepository {
	return NewMediaRepository(db)
}

func ProvideMediaTypeRepository(db *gorm.DB) *MediaTypeRepository {
	return NewMediaTypeRepository(db)
}

func ProvideStorageClient(cfg *config.Config) *Client {
	return NewClient(cfg)
}

func ProvideMediaService(
	mediaRepo *MediaRepository,
	mediaTypeRepo *MediaTypeRepository,
	storageClient *Client,
) *MediaService {
	return NewMediaService(mediaRepo, mediaTypeRepo, storageClient)
}