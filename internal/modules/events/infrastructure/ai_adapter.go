package infrastructure


import (
	"context"
	"fmt"
	"log"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/service"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/ai"
)

// ============================================================
// EVENTS AI ADAPTER
// ============================================================
//
// Thin wrapper around shared/ai.Client. No domain knowledge, no
// prompt construction, no parsing — just forwards two prompts and
// returns the raw response.

type EventAIAdapter struct {
	client *ai.Client
}

// NewEventAIAdapter wires a shared ai.Client into the events AIService.
func NewEventAIAdapter(client *ai.Client) service.AIService {
	return &EventAIAdapter{client: client}
}

func (a *EventAIAdapter) GenerateEventDraft(
	ctx context.Context,
	req service.GenerateEventDraftAIRequest,
) (string, error) {
	if !a.client.Enabled() {
		return "", ai.ErrDisabled
	}

	opts := make([]ai.ChatOption, 0, 2)
	if req.MaxTokens > 0 {
		opts = append(opts, ai.WithMaxTokens(req.MaxTokens))
	}
	if req.Temperature > 0 {
		opts = append(opts, ai.WithTemperature(req.Temperature))
	}

	raw, err := a.client.Chat(ctx, req.SystemPrompt, req.UserPrompt, opts...)
	if err != nil {
		return "", fmt.Errorf("event-ai chat failed: %w", err)
	}

	log.Printf("[event-ai] raw response chars=%d", len(raw))
	return raw, nil
}

var _ service.AIService = (*EventAIAdapter)(nil)