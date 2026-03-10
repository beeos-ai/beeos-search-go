//go:build integration

package beeos_test

import (
	"context"
	"os"
	"testing"
	"time"

	beeos "github.com/beeos-ai/beeos-search-go"
)

func getEnvOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func TestIntegration_FullFlow(t *testing.T) {
	baseURL := getEnvOrDefault("BEEOS_SEARCH_URL", "http://localhost:8099")
	mgmtKey := getEnvOrDefault("BEEOS_MANAGEMENT_KEY", "mgmt_test_token_2026")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Step 1: Health check
	t.Run("Healthz", func(t *testing.T) {
		client := beeos.NewClient(baseURL, "")
		if err := client.Healthz(ctx); err != nil {
			t.Fatalf("healthz failed: %v", err)
		}
		t.Log("healthz OK")
	})

	// Step 2: Admin login
	admin := beeos.NewAdminClient(baseURL, mgmtKey)

	t.Run("AdminLogin", func(t *testing.T) {
		if err := admin.Login(ctx); err != nil {
			t.Fatalf("admin login failed: %v", err)
		}
		t.Log("admin login OK")
	})

	// Step 3: Create API key
	var newKeyPlaintext string
	var newKeyID string

	t.Run("CreateKey", func(t *testing.T) {
		resp, err := admin.CreateKey(ctx, &beeos.CreateKeyRequest{
			Name:      "integration-test-key",
			RateLimit: 120,
		})
		if err != nil {
			t.Fatalf("create key failed: %v", err)
		}
		if resp.Key == "" {
			t.Fatal("expected plaintext key in response")
		}
		if resp.ID == "" {
			t.Fatal("expected key ID in response")
		}
		newKeyPlaintext = resp.Key
		newKeyID = resp.ID
		t.Logf("created key: prefix=%s id=%s", resp.KeyPrefix, resp.ID)
	})

	// Step 4: List keys — new key should appear
	t.Run("ListKeys", func(t *testing.T) {
		keys, err := admin.ListKeys(ctx)
		if err != nil {
			t.Fatalf("list keys failed: %v", err)
		}
		found := false
		for _, k := range keys {
			if k.ID == newKeyID {
				found = true
				if k.Name != "integration-test-key" {
					t.Errorf("expected name 'integration-test-key', got %q", k.Name)
				}
				if k.RateLimit != 120 {
					t.Errorf("expected rate_limit 120, got %d", k.RateLimit)
				}
				break
			}
		}
		if !found {
			t.Error("newly created key not found in list")
		}
		t.Logf("list keys returned %d keys", len(keys))
	})

	// Step 5: Update key
	t.Run("UpdateKey", func(t *testing.T) {
		err := admin.UpdateKey(ctx, newKeyID, &beeos.UpdateKeyRequest{
			Name:      "integration-test-key-updated",
			RateLimit: 200,
		})
		if err != nil {
			t.Fatalf("update key failed: %v", err)
		}
		t.Log("update key OK")
	})

	// Step 6: Search using the newly created key
	client := admin.NewClientFromAdmin(newKeyPlaintext)

	t.Run("Providers", func(t *testing.T) {
		providers, err := client.Providers(ctx)
		if err != nil {
			t.Fatalf("providers failed: %v", err)
		}
		if len(providers) == 0 {
			t.Fatal("expected at least 1 provider")
		}
		for _, p := range providers {
			t.Logf("provider: %s (%s)", p.ID, p.Name)
		}
	})

	t.Run("Search", func(t *testing.T) {
		resp, err := client.Search(ctx, &beeos.SearchRequest{
			Query: "golang testing best practices",
			Limit: 5,
		})
		if err != nil {
			t.Fatalf("search failed: %v", err)
		}
		if resp.Meta.ResultCount == 0 {
			t.Log("WARNING: search returned 0 results (provider might be down)")
		}
		t.Logf("search returned %d results in %dms", resp.Meta.ResultCount, resp.Meta.DurationMs)
		for i, r := range resp.Results {
			t.Logf("  %d. %s — %s", i+1, r.Title, r.URL)
		}
	})

	// Step 7: Fetch a URL
	t.Run("Fetch", func(t *testing.T) {
		resp, err := client.Fetch(ctx, &beeos.FetchRequest{
			URL: "https://go.dev",
		})
		if err != nil {
			t.Fatalf("fetch failed: %v", err)
		}
		if resp.Content == "" {
			t.Error("expected non-empty content")
		}
		content := resp.Content
		if len(content) > 100 {
			content = content[:100] + "..."
		}
		t.Logf("fetched %s: %s", resp.SourceURL, content)
	})

	// Step 8: Usage summary (should have at least 1 request now)
	t.Run("UsageSummary", func(t *testing.T) {
		// Wait briefly for async usage logger to flush
		time.Sleep(2 * time.Second)

		summary, err := admin.GetUsageSummary(ctx, &beeos.UsageSummaryOptions{Days: 1})
		if err != nil {
			t.Fatalf("usage summary failed: %v", err)
		}
		t.Logf("usage: total=%d today=%d week=%d month=%d",
			summary.TotalRequests, summary.TotalToday, summary.TotalWeek, summary.TotalMonth)
	})

	t.Run("ProviderStats", func(t *testing.T) {
		stats, err := admin.GetProviderStats(ctx, 1)
		if err != nil {
			t.Fatalf("provider stats failed: %v", err)
		}
		for _, s := range stats {
			t.Logf("provider %s: count=%d avg=%.1fms errors=%d",
				s.Provider, s.TotalCount, s.AvgDuration, s.ErrorCount)
		}
	})

	t.Run("RecentLogs", func(t *testing.T) {
		logs, err := admin.GetRecentLogs(ctx, &beeos.RecentLogsOptions{Limit: 10})
		if err != nil {
			t.Fatalf("recent logs failed: %v", err)
		}
		t.Logf("recent logs: %d entries", len(logs))
		for _, l := range logs {
			t.Logf("  %s %s provider=%s status=%d dur=%dms",
				l.Endpoint, l.Query, l.Provider, l.StatusCode, l.DurationMs)
		}
	})

	// Step 9: Revoke key
	t.Run("RevokeKey", func(t *testing.T) {
		if err := admin.RevokeKey(ctx, newKeyID); err != nil {
			t.Fatalf("revoke key failed: %v", err)
		}
		t.Log("revoke key OK")
	})

	// Step 10: Confirm revoked key can't search
	// The backend caches API keys for 30s, so wait for the cache to expire.
	t.Run("RevokedKeyRejected", func(t *testing.T) {
		t.Log("waiting 32s for API key cache to expire...")
		time.Sleep(32 * time.Second)

		_, err := client.Search(ctx, &beeos.SearchRequest{Query: "test"})
		if err == nil {
			t.Fatal("expected error when using revoked key")
		}
		apiErr, ok := err.(*beeos.APIError)
		if !ok {
			t.Fatalf("expected APIError, got %T: %v", err, err)
		}
		if apiErr.StatusCode != 401 {
			t.Errorf("expected 401, got %d", apiErr.StatusCode)
		}
		t.Logf("revoked key correctly rejected: %s", apiErr.Message)
	})

	t.Log("=== Integration test complete ===")
}
