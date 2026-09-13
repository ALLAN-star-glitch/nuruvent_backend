// internal/shared/ai/types.go

package ai

// ============================================================
// WIRE TYPES — OpenAI-compatible Chat Completions
// ============================================================
//
// These types mirror the JSON wire format used by OpenAI, OpenRouter,
// Groq, Together, and any other provider that implements the
// /v1/chat/completions endpoint.
//
// They are intentionally minimal — only the fields this project needs.

// Request is the body sent to the provider's chat completions endpoint.
type Request struct {
	Model          string            `json:"model"`
	Messages       []Message         `json:"messages"`
	Temperature    float64           `json:"temperature,omitempty"`
	MaxTokens      int               `json:"max_tokens,omitempty"`
	ResponseFormat map[string]string `json:"response_format,omitempty"`
}

// Message is a single turn in the conversation.
type Message struct {
	Role    string `json:"role"`    // "system" | "user" | "assistant"
	Content string `json:"content"`
}

// Response is the provider's reply.
type Response struct {
	ID      string       `json:"id"`
	Choices []Choice     `json:"choices"`
	Error   *RespError   `json:"error,omitempty"`
}

// Choice is a single completion candidate.
type Choice struct {
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// RespError is the provider-reported error.
type RespError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    int    `json:"code"`
}