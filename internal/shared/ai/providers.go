package ai

import (
	"github.com/google/wire"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/config"
)

// ProviderSet wires the shared AI client into the app graph.
//
// Register this set once in the top-level wire.Build(...). Every module
// that needs *ai.Client (team, events, ...) will be satisfied by it.
var ProviderSet = wire.NewSet(
	ProvideClient,
)

// ProvideClient is a thin wrapper around NewClient so Wire has a stable
// provider signature to bind. If construction grows (options, tracing,
// retries), extend it here without touching any module.
func ProvideClient(cfg *config.Config) *Client {
	return NewClient(cfg)
}