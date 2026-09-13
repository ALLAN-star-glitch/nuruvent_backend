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
	defaultTimeoutSeconds = 90
	defaultEndpoint       = "https://openrouter.ai/api/v1/chat/completions"
	defaultMaxTokens      = 1500
	defaultTemperature    = 0.7

	// Retry tuning for transient network failures.
	maxNetworkRetries     = 2
	networkRetryBackoffMs = 500
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
//
// Transient network failures (unexpected EOF, connection reset, timeouts)
// are retried up to maxNetworkRetries times with a short backoff. This
// absorbs the intermittent socket drops that OpenRouter's free tier
// produces under load.
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

	// ---- Attempt loop ----
	// Up to maxNetworkRetries+1 attempts. Only transient network errors
	// trigger a retry; HTTP-level errors (non-200) break the loop and are
	// returned as-is.
	var resp *http.Response
	var doErr error

	for attempt := 0; attempt <= maxNetworkRetries; attempt++ {
		// Build a fresh request every time. The previous body reader was
		// consumed by the failed attempt.
		httpReq, buildErr := http.NewRequestWithContext(
			ctx,
			http.MethodPost,
			c.endpoint,
			bytes.NewReader(jsonBody),
		)
		if buildErr != nil {
			return "", fmt.Errorf("failed to create request: %w", buildErr)
		}
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("HTTP-Referer", "https://nuruvent.com")
		httpReq.Header.Set("X-Title", "Nuruvent")

		resp, doErr = c.http.Do(httpReq)
		if doErr == nil {
			// Got a response — stop retrying, regardless of status code.
			// HTTP-level errors are handled below.
			break
		}

		if !isTransientNetworkError(doErr) {
			// Non-transient failure (DNS error, invalid URL, etc.).
			// No point retrying.
			return "", fmt.Errorf("ai provider request failed: %w", doErr)
		}

		if attempt < maxNetworkRetries {
			log.Printf(
				"⚠️ [ai] transient network error (attempt %d/%d): %v — retrying in %dms",
				attempt+1, maxNetworkRetries+1, doErr, networkRetryBackoffMs,
			)
			time.Sleep(time.Duration(networkRetryBackoffMs) * time.Millisecond)
		}
	}

	if doErr != nil {
		return "", fmt.Errorf(
			"ai provider request failed after %d attempts: %w",
			maxNetworkRetries+1, doErr,
		)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		msg := truncate(string(body), 300)
		log.Printf("❌ [ai] provider returned %d: %s", resp.StatusCode, msg)
		return "", fmt.Errorf("ai provider returned status %d: %s", resp.StatusCode, msg)
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

// ============================================================
// RETRY HELPERS
// ============================================================

// isTransientNetworkError reports whether err is a network-level failure
// that's worth retrying. Only genuine transport errors qualify — protocol
// errors (like TLS certificate validation failures) and application errors
// (like a 500 with a valid response body) do not.
//
// The strings checked here match what Go's net/http package and the
// underlying net.Conn produce for the failure modes we care about.
func isTransientNetworkError(err error) bool {
	if err == nil {
		return false
	}

	// Direct match on Go's sentinel.
	if errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}

	msg := err.Error()

	// String matches for wrapped errors that don't expose sentinels.
	// Order doesn't matter; any hit triggers a retry.
	transientFragments := []string{
		"unexpected EOF",
		"connection reset by peer",
		"connection refused",
		"i/o timeout",
		"TLS handshake timeout",
		"no such host",             // DNS flake
		"server misbehaving",       // DNS flake
		"temporary failure in name resolution",
		"network is unreachable",
		"broken pipe",
	}

	for _, frag := range transientFragments {
		if strings.Contains(msg, frag) {
			return true
		}
	}

	return false
}

// truncate returns s limited to n characters, appending "…" when shortened.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}