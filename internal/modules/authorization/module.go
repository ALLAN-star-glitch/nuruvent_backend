// internal/modules/authorization/module.go

package authorization

import (
	"context"
	"log"

	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/config"
	"gorm.io/gorm"
)

// Module represents the permissions module
type Module struct {
	enforcer *Enforcer
	service  *Service
}

// NewModule creates a new permissions module
func NewModule(db *gorm.DB, cfg *config.Config) (*Module, error) {
	// Initialize enforcer
	enforcer, err := InitEnforcer(db, cfg)
	if err != nil {
		return nil, err
	}

	// Initialize service
	service := NewService(enforcer)

	return &Module{
		enforcer: enforcer,
		service:  service,
	}, nil
}

// Init initializes the module
func (m *Module) Init(ctx context.Context) error {
	log.Println("Permissions module initialized")
	return nil
}

// GetEnforcer returns the enforcer instance
func (m *Module) GetEnforcer() *Enforcer {
	return m.enforcer
}

// GetService returns the permission service
func (m *Module) GetService() *Service {
	return m.service
}

// Close closes the module and cleans up resources
func (m *Module) Close() {
	if m.enforcer != nil {
		m.enforcer.Close()
	}
}