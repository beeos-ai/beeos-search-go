# beeos-search-go

Go SDK for the [BeeOS Search API](https://search.beeos.ai) — web search, content fetching, and admin management.

## Install

```bash
go get github.com/beeos-ai/beeos-search-go
```

## Quick Start

### Search & Fetch

```go
package main

import (
	"context"
	"fmt"
	"log"

	beeos "github.com/beeos-ai/beeos-search-go"
)

func main() {
	client := beeos.NewClient("https://search.beeos.ai", "bs_xxxxxxxxxxxx")
	ctx := context.Background()

	// Search
	resp, err := client.Search(ctx, &beeos.SearchRequest{
		Query: "BeeOS AI agent platform",
		Limit: 5,
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, r := range resp.Results {
		fmt.Printf("%s — %s\n", r.Title, r.URL)
	}

	// Fetch page content
	page, err := client.Fetch(ctx, &beeos.FetchRequest{
		URL: resp.Results[0].URL,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(page.Content)
}
```

### Admin (Key Management & Usage)

```go
admin := beeos.NewAdminClient("https://search.beeos.ai", "mgmt_xxxxxxxxxxxx")
ctx := context.Background()

// Create an API key
key, _ := admin.CreateKey(ctx, &beeos.CreateKeyRequest{
	Name:      "my-agent",
	RateLimit: 60,
})
fmt.Println("New key:", key.Key) // only shown once

// List keys
keys, _ := admin.ListKeys(ctx)

// Usage summary
summary, _ := admin.GetUsageSummary(ctx, &beeos.UsageSummaryOptions{Days: 7})
fmt.Printf("Requests this week: %d\n", summary.TotalWeek)
```

## API Reference

### Client

| Method | Description |
|---|---|
| `NewClient(baseURL, apiKey, ...Option)` | Create a search client |
| `Search(ctx, *SearchRequest)` | Web search with multi-provider support |
| `Fetch(ctx, *FetchRequest)` | Extract content from a URL |
| `Providers(ctx)` | List available search providers |
| `Healthz(ctx)` | Health check |

### AdminClient

| Method | Description |
|---|---|
| `NewAdminClient(baseURL, mgmtKey, ...Option)` | Create an admin client |
| `Login(ctx)` | Validate management key |
| `CreateKey(ctx, *CreateKeyRequest)` | Create a new API key |
| `ListKeys(ctx)` | List all API keys |
| `UpdateKey(ctx, keyID, *UpdateKeyRequest)` | Update key name or rate limit |
| `RevokeKey(ctx, keyID)` | Revoke an API key |
| `GetUsageSummary(ctx, *UsageSummaryOptions)` | Aggregated usage stats |
| `GetProviderStats(ctx, days)` | Per-provider statistics |
| `GetKeyUsage(ctx, keyID, *KeyUsageOptions)` | Per-key usage stats |
| `GetRecentLogs(ctx, *RecentLogsOptions)` | Recent usage logs |
| `NewClientFromAdmin(apiKey, ...Option)` | Create a Client from an admin-provisioned key |

### Options

```go
// Custom HTTP client
client := beeos.NewClient(url, key, beeos.WithHTTPClient(myHTTPClient))

// Custom timeout (default: 30s)
client := beeos.NewClient(url, key, beeos.WithTimeout(10*time.Second))
```

### Error Handling

API errors are returned as `*beeos.APIError` with `StatusCode`, `Code`, and `Message`:

```go
resp, err := client.Search(ctx, req)
if err != nil {
	var apiErr *beeos.APIError
	if errors.As(err, &apiErr) {
		fmt.Printf("API error %d: %s\n", apiErr.StatusCode, apiErr.Message)
	}
}
```

## Environment Variables

| Variable | Description |
|---|---|
| `BEEOS_SEARCH_URL` | API base URL (default: `http://localhost:8099`) |
| `BEEOS_API_KEY` | Search API key (`bs_...`) |
| `BEEOS_MANAGEMENT_KEY` | Admin management key (`mgmt_...`) |

## License

MIT
