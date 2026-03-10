# BeeOS Search Go SDK — Agent Context

## Overview

Go client library for the BeeOS Search API. Provides two clients:

- **`Client`** — search, fetch, and provider listing (uses API Key auth)
- **`AdminClient`** — API key management and usage analytics (uses Management Key auth)

Module path: `github.com/beeos-ai/beeos-search-go`  
Package name: `beeos`

## File Map

| File | Purpose |
|------|---------|
| `client.go` | `Client` struct — `Search()`, `Fetch()`, `Providers()`, `Healthz()` + shared HTTP helper |
| `admin.go` | `AdminClient` struct — `CreateKey()`, `ListKeys()`, `UpdateKey()`, `RevokeKey()`, usage queries, `Login()` |
| `types.go` | All request/response types, `APIError` |
| `client_test.go` | Unit tests for Client |
| `admin_test.go` | Unit tests for AdminClient |
| `integration_test.go` | Integration tests (require running backend) |
| `examples/` | Usage examples |

## API Coverage

### Client (API Key)

| Method | Backend Endpoint |
|--------|-----------------|
| `Search()` | `POST /v1/search` |
| `Fetch()` | `POST /v1/fetch` |
| `Providers()` | `GET /v1/providers` |
| `Healthz()` | `GET /healthz` |

### AdminClient (Management Key)

| Method | Backend Endpoint |
|--------|-----------------|
| `CreateKey()` | `POST /admin/keys` |
| `ListKeys()` | `GET /admin/keys` |
| `UpdateKey()` | `PATCH /admin/keys/:id` |
| `RevokeKey()` | `DELETE /admin/keys/:id` |
| `GetUsageSummary()` | `GET /admin/usage/summary` |
| `GetProviderStats()` | `GET /admin/usage/providers` |
| `GetKeyUsage()` | `GET /admin/usage/keys/:id` |
| `GetRecentLogs()` | `GET /admin/usage/recent` |
| `Login()` | `POST /admin/auth/login` |

## Conventions

- Zero external dependencies — stdlib only (`net/http`, `encoding/json`)
- All methods accept `context.Context` as first argument
- Errors for HTTP 4xx/5xx are returned as `*APIError` (type-assertable)
- `Option` pattern for configuring clients: `WithHTTPClient()`, `WithTimeout()`
- `WaitForReady()` helper polls `/healthz` until the backend is up

## Testing

```bash
go test -v ./...                              # unit tests
go test -v -tags=integration -run Integration  # integration tests (needs running backend)
```
