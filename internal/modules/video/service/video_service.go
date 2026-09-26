// internal/modules/video/service/video_service.go

package service

// videoService is the concrete implementation of Service.
//
// Each use case lives in its own file (connect.go, callback.go,
// disconnect.go, create_meeting.go, list.go, refresh.go). This file
// only holds the type and its constructor.
type videoService struct {
	deps Dependencies
}

// New constructs the video service.
//
// Fails fast when a required dependency is missing — a zero-value
// dep is almost always a wiring bug, and failing at startup is
// better than a nil-pointer panic at first request.
func New(deps Dependencies) (Service, error) {
	if deps.UnitOfWork == nil {
		return nil, errMissingDependency("UnitOfWork")
	}
	if deps.Connections == nil {
		return nil, errMissingDependency("Connections")
	}
	if deps.OAuthStates == nil {
		return nil, errMissingDependency("OAuthStates")
	}
	if deps.Meetings == nil {
		return nil, errMissingDependency("Meetings")
	}
	if deps.Clients == nil {
		return nil, errMissingDependency("Clients")
	}
	if deps.IDs == nil {
		return nil, errMissingDependency("IDs")
	}
	if deps.Clock == nil {
		return nil, errMissingDependency("Clock")
	}
	return &videoService{deps: deps}, nil
}