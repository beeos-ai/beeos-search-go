// Package beeos provides a Go SDK for the BeeOS Search API.
//
// Two clients are provided:
//   - Client: uses an API Key for search and fetch operations.
//   - AdminClient: uses a Management API Key for API key CRUD and usage queries.
package beeos

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client provides access to the BeeOS Search API (search, fetch, providers).
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// Option configures a Client or AdminClient.
type Option func(*clientOptions)

type clientOptions struct {
	httpClient *http.Client
	timeout    time.Duration
}

// WithHTTPClient sets a custom http.Client.
func WithHTTPClient(c *http.Client) Option {
	return func(o *clientOptions) { o.httpClient = c }
}

// WithTimeout sets the HTTP request timeout. Default is 30s.
func WithTimeout(d time.Duration) Option {
	return func(o *clientOptions) { o.timeout = d }
}

func applyOptions(opts []Option) *clientOptions {
	o := &clientOptions{timeout: 30 * time.Second}
	for _, fn := range opts {
		fn(o)
	}
	if o.httpClient == nil {
		o.httpClient = &http.Client{Timeout: o.timeout}
	}
	return o
}

// NewClient creates a new Client for search and fetch operations.
//
//	client := beeos.NewClient("https://search.beeos.ai", "bs_xxxxxxxxxxxx")
func NewClient(baseURL, apiKey string, opts ...Option) *Client {
	o := applyOptions(opts)
	return &Client{
		baseURL:    baseURL,
		apiKey:     apiKey,
		httpClient: o.httpClient,
	}
}

// Search performs a search query.
func (c *Client) Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	var resp SearchResponse
	if err := c.doJSON(ctx, http.MethodPost, "/v1/search", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Fetch retrieves the content of a URL.
func (c *Client) Fetch(ctx context.Context, req *FetchRequest) (*FetchResponse, error) {
	var resp FetchResponse
	if err := c.doJSON(ctx, http.MethodPost, "/v1/fetch", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Providers lists all registered search providers.
func (c *Client) Providers(ctx context.Context) ([]ProviderInfo, error) {
	var resp struct {
		Providers []ProviderInfo `json:"providers"`
	}
	if err := c.doJSON(ctx, http.MethodGet, "/v1/providers", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Providers, nil
}

// Healthz checks the service health.
func (c *Client) Healthz(ctx context.Context) error {
	return c.doJSON(ctx, http.MethodGet, "/healthz", nil, nil)
}

func (c *Client) doJSON(ctx context.Context, method, path string, body, result any) error {
	return doJSONRequest(ctx, c.httpClient, c.baseURL, c.apiKey, method, path, body, result)
}

// --- shared HTTP helper ---

func doJSONRequest(ctx context.Context, hc *http.Client, baseURL, token, method, path string, body, result any) error {
	// Split path from query string before joining, because url.JoinPath
	// escapes '?' in path segments.
	queryStr := ""
	if idx := strings.Index(path, "?"); idx != -1 {
		queryStr = path[idx:]
		path = path[:idx]
	}
	u, err := url.JoinPath(baseURL, path)
	if err != nil {
		return fmt.Errorf("beeos: invalid URL: %w", err)
	}
	u += queryStr

	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("beeos: marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, u, bodyReader)
	if err != nil {
		return fmt.Errorf("beeos: create request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := hc.Do(req)
	if err != nil {
		return fmt.Errorf("beeos: request failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("beeos: read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		apiErr := &APIError{StatusCode: resp.StatusCode}
		var wrapper struct {
			Error APIError `json:"error"`
		}
		if json.Unmarshal(data, &wrapper) == nil && wrapper.Error.Code != "" {
			apiErr.Code = wrapper.Error.Code
			apiErr.Message = wrapper.Error.Message
		} else if json.Unmarshal(data, apiErr) != nil || apiErr.Code == "" {
			apiErr.Code = http.StatusText(resp.StatusCode)
			apiErr.Message = string(data)
		}
		return apiErr
	}

	if result != nil && len(data) > 0 {
		if err := json.Unmarshal(data, result); err != nil {
			return fmt.Errorf("beeos: decode response: %w", err)
		}
	}
	return nil
}
