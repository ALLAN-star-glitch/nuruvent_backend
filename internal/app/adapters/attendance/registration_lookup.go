package attendance

import (
	"context"
	"errors"
	"fmt"

	attendanceDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
	registrationDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/registrationdomain"
	"gorm.io/gorm"
)

// RegistrationLookupAdapter resolves the user behind a registration
// by reading the registration repository directly.
//
// It intentionally does NOT go through the registration service —
// the service already depends on the attendance service, so calling
// it from here would create a dependency cycle at wire time.
type RegistrationLookupAdapter struct {
	eventRegs registrationDomain.EventRegistrationRepository
}

func NewRegistrationLookupAdapter(
	eventRegs registrationDomain.EventRegistrationRepository,
) attendanceDomain.RegistrationLookup {
	return &RegistrationLookupAdapter{eventRegs: eventRegs}
}

func (a *RegistrationLookupAdapter) GetUserIDByRegistrationID(
	ctx context.Context,
	registrationID string,
) (string, error) {
	if registrationID == "" {
		return "", nil
	}

	reg, err := a.eventRegs.FindByID(ctx, registrationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil
		}
		return "", fmt.Errorf("find registration: %w", err)
	}
	if reg == nil || reg.Registration == nil {
		return "", nil
	}
	return reg.Registration.UserID, nil
}