// internal/modules/account/module.go

package account

import (
	"context"
	"log"

	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/config"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/database"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/account/accountrepo"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/account/accountservice"
)

type Module struct {
	service accountservice.Service
}

func NewModule(
	cfg *config.Config,
) *Module {
	db := database.GetDB()
	repo := accountrepo.NewRepository(db)
	svc := accountservice.NewService(repo)

	return &Module{
		service: svc,
	}
}

// GetService returns the account service
func (m *Module) GetService() accountservice.Service {
	return m.service
}

func (m *Module) Init(ctx context.Context) error {
	log.Println("Account module initialized")
	return nil
}

func (m *Module) Close() {
	log.Println("Account module closed")
}