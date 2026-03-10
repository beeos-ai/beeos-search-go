package beeos

import "time"

// SearchRequest is the request body for POST /v1/search.
type SearchRequest struct {
	Query          string   `json:"query"`
	Limit          int      `json:"limit,omitempty"`
	Region         string   `json:"region,omitempty"`
	Providers      []string `json:"providers,omitempty"`
	IncludeContent bool     `json:"include_content,omitempty"`
}

// SearchResult is a single search result returned by the API.
type SearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
	Content string `json:"content,omitempty"`
	Source  string `json:"source,omitempty"`
	Date    string `json:"date,omitempty"`
}

// SearchMeta contains metadata about the search response.
type SearchMeta struct {
	Query           string           `json:"query"`
	ProvidersUsed   []string         `json:"providers_used,omitempty"`
	ProvidersFailed []ProviderError  `json:"providers_failed,omitempty"`
	ResultCount     int              `json:"result_count"`
	Cached          bool             `json:"cached"`
	DurationMs      int64            `json:"duration_ms"`
}

// ProviderError describes a provider that failed during search.
type ProviderError struct {
	Provider string `json:"provider"`
	Error    string `json:"error"`
}

// SearchResponse is the response from POST /v1/search.
type SearchResponse struct {
	Results []SearchResult `json:"results"`
	Meta    SearchMeta     `json:"meta"`
}

// FetchRequest is the request body for POST /v1/fetch.
type FetchRequest struct {
	URL string `json:"url"`
}

// FetchResponse is the response from POST /v1/fetch.
type FetchResponse struct {
	Content     string `json:"content"`
	ContentType string `json:"content_type"`
	Title       string `json:"title,omitempty"`
	SourceURL   string `json:"source_url"`
}

// ProviderInfo describes a registered search provider.
type ProviderInfo struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Regions []string `json:"regions"`
	Status  string `json:"status"`
}

// APIKey represents an API key record (without the secret).
type APIKey struct {
	ID         string     `json:"id"`
	KeyPrefix  string     `json:"key_prefix"`
	Name       string     `json:"name"`
	CreatedAt  time.Time  `json:"created_at"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	RateLimit  int        `json:"rate_limit"`
}

// CreateKeyRequest is the request body for creating a new API key.
type CreateKeyRequest struct {
	Name      string `json:"name"`
	RateLimit int    `json:"rate_limit,omitempty"`
}

// CreateKeyResponse is returned when a new API key is created.
// The PlaintextKey is only available at creation time.
type CreateKeyResponse struct {
	APIKey
	Key string `json:"key"`
}

// UpdateKeyRequest is the request body for updating an API key.
type UpdateKeyRequest struct {
	Name      string `json:"name,omitempty"`
	RateLimit int    `json:"rate_limit,omitempty"`
}

// UsageSummary contains aggregated usage statistics.
type UsageSummary struct {
	TotalRequests int64           `json:"total_requests"`
	TotalToday    int64           `json:"total_today"`
	TotalWeek     int64           `json:"total_week"`
	TotalMonth    int64           `json:"total_month"`
	ByProvider    []ProviderStats `json:"by_provider"`
	Daily         []DailyStat     `json:"daily"`
}

// ProviderStats contains per-provider usage statistics.
type ProviderStats struct {
	Provider    string  `json:"provider"`
	TotalCount  int64   `json:"total_count"`
	AvgDuration float64 `json:"avg_duration_ms"`
	ErrorCount  int64   `json:"error_count"`
}

// DailyStat contains per-day usage data.
type DailyStat struct {
	Date       string `json:"date"`
	Provider   string `json:"provider"`
	TotalCount int64  `json:"total_count"`
}

// UsageLog is a single usage log entry.
type UsageLog struct {
	ID          int64     `json:"id"`
	APIKeyID    string    `json:"api_key_id"`
	Endpoint    string    `json:"endpoint"`
	Provider    string    `json:"provider"`
	Query       string    `json:"query"`
	Region      string    `json:"region"`
	StatusCode  int       `json:"status_code"`
	ResultCount int       `json:"result_count"`
	DurationMs  int       `json:"duration_ms"`
	CreatedAt   time.Time `json:"created_at"`
}

// APIError is returned when the API returns a non-2xx status.
type APIError struct {
	StatusCode int    `json:"-"`
	Code       string `json:"code"`
	Message    string `json:"message"`
}

func (e *APIError) Error() string {
	return e.Code + ": " + e.Message
}
