package beeos

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// AdminClient provides access to the BeeOS Search Admin API.
// It uses a Management API Key (AdminToken) for authentication.
type AdminClient struct {
	baseURL        string
	managementKey  string
	httpClient     *http.Client
}

// NewAdminClient creates a new AdminClient for API key management and usage queries.
//
//	admin := beeos.NewAdminClient("https://search.beeos.ai", "mgmt_xxxxxxxxxxxx")
func NewAdminClient(baseURL, managementKey string, opts ...Option) *AdminClient {
	o := applyOptions(opts)
	return &AdminClient{
		baseURL:       baseURL,
		managementKey: managementKey,
		httpClient:    o.httpClient,
	}
}

// --- API Key Management ---

// CreateKey creates a new API key. The returned CreateKeyResponse.Key
// contains the plaintext key, which is only available at creation time.
func (a *AdminClient) CreateKey(ctx context.Context, req *CreateKeyRequest) (*CreateKeyResponse, error) {
	var resp CreateKeyResponse
	if err := a.doJSON(ctx, http.MethodPost, "/admin/keys", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListKeys returns all API keys.
func (a *AdminClient) ListKeys(ctx context.Context) ([]APIKey, error) {
	var resp struct {
		Keys []APIKey `json:"keys"`
	}
	if err := a.doJSON(ctx, http.MethodGet, "/admin/keys", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Keys, nil
}

// UpdateKey updates an existing API key's name or rate limit.
func (a *AdminClient) UpdateKey(ctx context.Context, keyID string, req *UpdateKeyRequest) error {
	return a.doJSON(ctx, http.MethodPatch, "/admin/keys/"+keyID, req, nil)
}

// RevokeKey revokes an API key by its ID. Revoked keys can no longer be used.
func (a *AdminClient) RevokeKey(ctx context.Context, keyID string) error {
	return a.doJSON(ctx, http.MethodDelete, "/admin/keys/"+keyID, nil, nil)
}

// --- Usage & Analytics ---

// UsageSummaryOptions configures the usage summary query.
type UsageSummaryOptions struct {
	Days int // Number of days to include. Default: 30
}

// GetUsageSummary returns aggregated usage statistics.
func (a *AdminClient) GetUsageSummary(ctx context.Context, opts *UsageSummaryOptions) (*UsageSummary, error) {
	path := "/admin/usage/summary"
	if opts != nil && opts.Days > 0 {
		path = fmt.Sprintf("%s?days=%d", path, opts.Days)
	}
	var resp UsageSummary
	if err := a.doJSON(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetProviderStats returns per-provider usage statistics.
func (a *AdminClient) GetProviderStats(ctx context.Context, days int) ([]ProviderStats, error) {
	path := "/admin/usage/providers"
	if days > 0 {
		path = fmt.Sprintf("%s?days=%d", path, days)
	}
	var resp struct {
		Providers []ProviderStats `json:"providers"`
	}
	if err := a.doJSON(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Providers, nil
}

// KeyUsageOptions configures the per-key usage query.
type KeyUsageOptions struct {
	Days int
}

// GetKeyUsage returns usage statistics for a specific API key.
func (a *AdminClient) GetKeyUsage(ctx context.Context, keyID string, opts *KeyUsageOptions) ([]ProviderStats, error) {
	path := "/admin/usage/keys/" + keyID
	if opts != nil && opts.Days > 0 {
		path = fmt.Sprintf("%s?days=%d", path, opts.Days)
	}
	var resp struct {
		Usage []ProviderStats `json:"usage"`
	}
	if err := a.doJSON(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Usage, nil
}

// RecentLogsOptions configures the recent logs query.
type RecentLogsOptions struct {
	Limit  int
	Offset int
}

// GetRecentLogs returns recent usage log entries.
func (a *AdminClient) GetRecentLogs(ctx context.Context, opts *RecentLogsOptions) ([]UsageLog, error) {
	path := "/admin/usage/recent"
	if opts != nil {
		sep := "?"
		if opts.Limit > 0 {
			path += fmt.Sprintf("%slimit=%d", sep, opts.Limit)
			sep = "&"
		}
		if opts.Offset > 0 {
			path += fmt.Sprintf("%soffset=%d", sep, opts.Offset)
		}
	}
	var resp struct {
		Logs []UsageLog `json:"logs"`
	}
	if err := a.doJSON(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Logs, nil
}

// Login validates the management API key against the server.
// Returns nil on success or an error if the key is invalid.
func (a *AdminClient) Login(ctx context.Context) error {
	return a.doJSON(ctx, http.MethodPost, "/admin/auth/login", nil, nil)
}

// NewClientFromAdmin creates a regular Client using a key created via CreateKey.
// This is a convenience for provisioning and immediately using a new API key.
func (a *AdminClient) NewClientFromAdmin(apiKey string, opts ...Option) *Client {
	return NewClient(a.baseURL, apiKey, opts...)
}

func (a *AdminClient) doJSON(ctx context.Context, method, path string, body, result any) error {
	return doJSONRequest(ctx, a.httpClient, a.baseURL, a.managementKey, method, path, body, result)
}

// --- Helper: wait for service readiness ---

// WaitForReady polls the /healthz endpoint until the service is ready or the context is cancelled.
func WaitForReady(ctx context.Context, baseURL string, interval time.Duration) error {
	c := NewClient(baseURL, "")
	for {
		if err := c.Healthz(ctx); err == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(interval):
		}
	}
}
