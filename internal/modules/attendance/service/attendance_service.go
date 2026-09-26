// internal/modules/attendance/service/attendance_service.go

package service

// attendanceService is the concrete implementation of Service.
//
// Each use case lives in its own file (register_attendee.go,
// upsert_session.go, etc.). This file only holds the type and its
// constructor.
type attendanceService struct {
	deps Dependencies
}

// NewService constructs the attendance service. Fails fast if a
// required dependency is missing or the derivation policy is invalid.
func NewService(deps Dependencies) (Service, error) {
	if deps.UnitOfWork == nil {
		return nil, errMissingDependency("UnitOfWork")
	}
	if deps.Clock == nil {
		return nil, errMissingDependency("Clock")
	}
	if deps.IDs == nil {
		return nil, errMissingDependency("IDs")
	}
	if deps.TokenGenerator == nil {
		return nil, errMissingDependency("TokenGenerator")
	}
	if deps.Publisher == nil {
		return nil, errMissingDependency("Publisher")
	}
	if err := deps.DerivationPolicy.Validate(); err != nil {
		return nil, err
	}
	if deps.JoinTokenGrace <= 0 {
		return nil, errMissingDependency("JoinTokenGrace")
	}

	return &attendanceService{deps: deps}, nil
}