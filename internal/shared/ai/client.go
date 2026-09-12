// internal/shared/ai/client.go

package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/config"
)

// ============================================================
// SHARED AI CLIENT
// ============================================================
//
// Client is a minimal, provider-agnostic wrapper around an
// OpenAI-compatible Chat Completions endpoint (OpenRouter, OpenAI, Groq,
// Together, etc.).
//
// Callers supply a system prompt and a user prompt. The client returns the
// assistant's raw response string. Prompt construction and response
// parsing remain the caller's responsibility.
//
// This package is transport-only. It does not know about events, teams,
// invitations, or any domain concept.

const (
	defaultModel          = "meta-llama/llama-3.1-8b-instruct:free"
	defaultTimeoutSeconds = 30
	defaultEndpoint       = "https://openrouter.ai/api/v1/chat/completions"
	defaultMaxTokens      = 1500
	defaultTemperature    = 0.7
)

// ErrDisabled is returned when the client was constructed without an API key.
var ErrDisabled = errors.New("ai client is disabled (missing api_key)")

// Client sends chat completions requests to the configured provider.
type Client struct {
	apiKey   string
	model    string
	endpoint string
	enabled  bool
	http     *http.Client
}

// NewClient constructs an AI client from configuration.
//
// If cfg.OpenRouter.APIKey is empty, the returned client is disabled and
// every call to Chat returns ErrDisabled. Callers should check Enabled()
// before relying on the client.
func NewClient(cfg *config.Config) *Client {
	apiKey := strings.TrimSpace(cfg.OpenRouter.APIKey)
	if apiKey == "" {
		log.Println("⚠️  [ai] OpenRouter API key not provided — AI features disabled")
		return &Client{enabled: false}
	}

	model := strings.TrimSpace(cfg.OpenRouter.Model)
	if model == "" {
		model = defaultModel
	}

	log.Printf("✅ [ai] client initialized (model=%s, timeout=%ds)", model, defaultTimeoutSeconds)

	return &Client{
		apiKey:   apiKey,
		model:    model,
		endpoint: defaultEndpoint,
		enabled:  true,
		http: &http.Client{
			Timeout: defaultTimeoutSeconds * time.Second,
		},
	}
}

// Enabled reports whether the client can make calls.
func (c *Client) Enabled() bool {
	return c != nil && c.enabled
}

// Model returns the configured model identifier.
func (c *Client) Model() string {
	if c == nil {
		return ""
	}
	return c.model
}

// ============================================================
// CHAT OPTIONS
// ============================================================

// ChatOption customizes a single Chat call.
type ChatOption func(*Request)

// WithMaxTokens overrides the default max_tokens for one call.
func WithMaxTokens(n int) ChatOption {
	return func(r *Request) {
		if n > 0 {
			r.MaxTokens = n
		}
	}
}

// WithTemperature overrides the default temperature for one call.
func WithTemperature(t float64) ChatOption {
	return func(r *Request) {
		if t >= 0 {
			r.Temperature = t
		}
	}
}

// WithJSONMode forces response_format=json_object. This is the
// default; expose it for callers that globally default to plain text.
func WithJSONMode() ChatOption {
	return func(r *Request) {
		r.ResponseFormat = map[string]string{"type": "json_object"}
	}
}

// WithPlainText disables JSON mode so the model can return freeform text.
func WithPlainText() ChatOption {
	return func(r *Request) {
		r.ResponseFormat = nil
	}
}

// ============================================================
// CHAT
// ============================================================

// Chat sends a two-message conversation (system + user) and returns the
// assistant's response content.
//
// Returns ErrDisabled if the client is disabled.
// Returns an error if the provider responds with a non-200 status or if the
// response cannot be parsed.
func (c *Client) Chat(
	ctx context.Context,
	system, user string,
	opts ...ChatOption,
) (string, error) {
	if !c.Enabled() {
		return "", ErrDisabled
	}
	if strings.TrimSpace(user) == "" {
		return "", errors.New("user prompt is required")
	}

	messages := make([]Message, 0, 2)
	if strings.TrimSpace(system) != "" {
		messages = append(messages, Message{Role: "system", Content: system})
	}
	messages = append(messages, Message{Role: "user", Content: user})

	reqBody := Request{
		Model:          c.model,
		Messages:       messages,
		Temperature:    defaultTemperature,
		MaxTokens:      defaultMaxTokens,
		ResponseFormat: map[string]string{"type": "json_object"},
	}
	for _, opt := range opts {
		opt(&reqBody)
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("HTTP-Referer", "https://nuruvent.com")
	httpReq.Header.Set("X-Title", "Nuruvent")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("ai provider request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("❌ [ai] provider returned %d: %s", resp.StatusCode, truncate(string(body), 200))
		return "", fmt.Errorf("ai provider returned status %d", resp.StatusCode)
	}

	var parsed Response
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("failed to parse provider response: %w", err)
	}

	if parsed.Error != nil {
		return "", fmt.Errorf("ai provider error: %s", parsed.Error.Message)
	}

	if len(parsed.Choices) == 0 {
		return "", errors.New("ai provider returned no choices")
	}

	return parsed.Choices[0].Message.Content, nil
}

// truncate returns s limited to n characters, appending "…" when shortened.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}