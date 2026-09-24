// internal/modules/payment/infrastructure/providers/intasend/client.go

package intasend

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/config"
)

// client is a thin HTTP wrapper around IntaSend's API.
//
// IntaSend has two auth modes:
//
//   - Authenticated requests (POST /payment/collection/, /payment/status/,
//     /payment/refund/) use the SECRET key in the Authorization header.
//     Use do() for these.
//
//   - Public requests (POST /checkout/) use the PUBLISHABLE key in the
//     JSON body and no Authorization header. Use doPublic() for these.
type client struct {
	baseURL    string
	secretKey  string
	httpClient *http.Client
}

func newClient(cfg config.IntaSendConfig) *client {
	return &client{
		baseURL:   cfg.BaseURL,
		secretKey: cfg.SecretKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// do performs an authenticated request and returns the raw response body.
//
// Uses the secret key as a Bearer token in the Authorization header.
// Non-2xx responses are converted to typed errors.
func (c *client) do(ctx context.Context, method, path string, body any) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal body: %w", err)
		}
		reqBody = bytes.NewReader(buf)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.secretKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http call: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return respBody, nil
	}

	return nil, translateHTTPError(resp.StatusCode, respBody)
}

// doPublic performs an unauthenticated request using the publishable
// key in the JSON body — no Authorization header.
//
// Used for the Checkout Link API (/checkout/), which accepts the
// publishable key in the body and does not need a Bearer token.
func (c *client) doPublic(ctx context.Context, method, path string, body any) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal body: %w", err)
		}
		reqBody = bytes.NewReader(buf)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	// Deliberately no Authorization header — this is a public endpoint.

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http call: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return respBody, nil
	}

	return nil, translateHTTPError(resp.StatusCode, respBody)
}