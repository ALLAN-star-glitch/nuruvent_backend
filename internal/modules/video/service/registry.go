// internal/modules/video/service/registry.go

package service

import (
	"fmt"

	videodomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/videodomain"
)

// clientRegistry is the concrete implementation of
// videodomain.ClientRegistry.
//
// It holds the clients registered at startup. Registration is done
// once during wiring; the registry is read-only after that.
type clientRegistry struct {
	clients map[videodomain.Platform]videodomain.ProviderClient
}

// NewClientRegistry constructs a registry from a map of clients.
//
// Clients with a nil value are ignored — allows the wiring to pass
// nil for platforms that aren't configured without polluting the map.
func NewClientRegistry(
	clients map[videodomain.Platform]videodomain.ProviderClient,
) videodomain.ClientRegistry {
	clean := make(map[videodomain.Platform]videodomain.ProviderClient, len(clients))
	for platform, client := range clients {
		if client == nil {
			continue
		}
		clean[platform] = client
	}
	return &clientRegistry{clients: clean}
}

// For returns the client registered for the platform.
func (r *clientRegistry) For(platform videodomain.Platform) (videodomain.ProviderClient, error) {
	client, ok := r.clients[platform]
	if !ok {
		return nil, fmt.Errorf("%w: %s", videodomain.ErrUnsupportedPlatform, platform)
	}
	return client, nil
}

// OAuthFor returns a client that implements OAuthProvider.
func (r *clientRegistry) OAuthFor(platform videodomain.Platform) (videodomain.OAuthProvider, error) {
	client, err := r.For(platform)
	if err != nil {
		return nil, err
	}
	oauth, ok := client.(videodomain.OAuthProvider)
	if !ok {
		return nil, fmt.Errorf("%w: platform %s does not support OAuth",
			videodomain.ErrCapabilityMissing, platform)
	}
	return oauth, nil
}

// ProvisionerFor returns a client that implements MeetingProvisioner.
func (r *clientRegistry) ProvisionerFor(platform videodomain.Platform) (videodomain.MeetingProvisioner, error) {
	client, err := r.For(platform)
	if err != nil {
		return nil, err
	}
	provisioner, ok := client.(videodomain.MeetingProvisioner)
	if !ok {
		return nil, fmt.Errorf("%w: platform %s cannot provision meetings",
			videodomain.ErrCapabilityMissing, platform)
	}
	return provisioner, nil
}

// MinterFor returns a client that implements RoomParticipantMinter.
func (r *clientRegistry) MinterFor(platform videodomain.Platform) (videodomain.RoomParticipantMinter, error) {
	client, err := r.For(platform)
	if err != nil {
		return nil, err
	}
	minter, ok := client.(videodomain.RoomParticipantMinter)
	if !ok {
		return nil, fmt.Errorf("%w: platform %s does not support participant tokens",
			videodomain.ErrCapabilityMissing, platform)
	}
	return minter, nil
}

// compile-time assertion
var _ videodomain.ClientRegistry = (*clientRegistry)(nil)