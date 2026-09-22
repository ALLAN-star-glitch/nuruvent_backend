package flutterwave

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

// client is a thin HTTP wrapper around Flutterwave's v3 API.
type client struct {
	baseURL    string
	secretKey  string
	httpClient *http.Client
}

func newClient(cfg config.FlutterwaveConfig) *client {
	return &client{
		baseURL:   cfg.BaseURL,
		secretKey: cfg.SecretKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// do performs an authenticated request and returns the raw response body.
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