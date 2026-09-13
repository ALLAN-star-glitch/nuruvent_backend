package service

import "context"

// ============================================================
// EVENTS AI PORT
// ============================================================
//
// AIService is the outbound port the events module uses to reach an
// LLM. The adapter behind it is transport-only: it forwards the two
// prompts and returns the raw assistant string. Parsing, validation,
// auto-correction, and retry all live in the service layer
// (see ai_generate_draft.go).
//
// This shape is intentional. The retry flow in the design doc (§7.6)
// requires rewriting the prompt after a validation failure — which
// means the *service* must own prompt construction, not the adapter.

type AIService interface {
	// GenerateEventDraft sends the given prompts and returns the raw
	// assistant response. Returns ai.ErrDisabled if the underlying
	// client has no API key configured.
	GenerateEventDraft(ctx context.Context, req GenerateEventDraftAIRequest) (string, error)
}

// GenerateEventDraftAIRequest carries the two prompts plus per-call
// generation parameters.
type GenerateEventDraftAIRequest struct {
	SystemPrompt string
	UserPrompt   string
	MaxTokens    int     // 0 → adapter default
	Temperature  float64 // 0 → adapter default
}