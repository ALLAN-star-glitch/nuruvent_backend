package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/registrationdomain"
	"gorm.io/gorm"
)

type RegistrationNumberGenerator struct {
	db *gorm.DB
}

// Constructor returns the concrete type, not the interface.
func NewRegistrationNumberGenerator(db *gorm.DB) *RegistrationNumberGenerator {
	return &RegistrationNumberGenerator{db: db}
}

func (g *RegistrationNumberGenerator) Next(ctx context.Context) (string, error) {
	var seq int64
	err := g.db.WithContext(ctx).
		Raw("SELECT nextval('registration_number_seq')").
		Scan(&seq).Error
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("REG-%d-%06d", time.Now().Year(), seq), nil
}

// Compile-time assertion that this type satisfies the domain port.
var _ registrationdomain.RegistrationNumberGenerator = (*RegistrationNumberGenerator)(nil)