package main

import (
	"context"
	"fmt"
	"log"
	"os"

	beeos "github.com/beeos-ai/beeos-search-go"
)

func main() {
	baseURL := os.Getenv("BEEOS_SEARCH_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8099"
	}
	mgmtKey := os.Getenv("BEEOS_MANAGEMENT_KEY")
	if mgmtKey == "" {
		log.Fatal("BEEOS_MANAGEMENT_KEY is required")
	}

	admin := beeos.NewAdminClient(baseURL, mgmtKey)
	ctx := context.Background()

	// Validate management key
	if err := admin.Login(ctx); err != nil {
		log.Fatalf("Login failed: %v", err)
	}
	fmt.Println("Management key validated.")

	// Create a new API key
	newKey, err := admin.CreateKey(ctx, &beeos.CreateKeyRequest{
		Name:      "demo-key",
		RateLimit: 60,
	})
	if err != nil {
		log.Fatalf("CreateKey failed: %v", err)
	}
	fmt.Printf("Created key: %s (prefix: %s)\n", newKey.Key, newKey.KeyPrefix)
	fmt.Println("⚠ Save the key above — it won't be shown again.")

	// List all keys
	keys, err := admin.ListKeys(ctx)
	if err != nil {
		log.Fatalf("ListKeys failed: %v", err)
	}
	fmt.Printf("\nAll keys (%d):\n", len(keys))
	for _, k := range keys {
		status := "active"
		if k.RevokedAt != nil {
			status = "revoked"
		}
		fmt.Printf("  - %s  %s  [%s]  rate_limit=%d\n", k.KeyPrefix, k.Name, status, k.RateLimit)
	}

	// Use the new key to search
	client := admin.NewClientFromAdmin(newKey.Key)
	resp, err := client.Search(ctx, &beeos.SearchRequest{Query: "test query", Limit: 3})
	if err != nil {
		log.Printf("Search with new key failed: %v", err)
	} else {
		fmt.Printf("\nSearch with new key returned %d results.\n", resp.Meta.ResultCount)
	}

	// Check usage
	summary, err := admin.GetUsageSummary(ctx, &beeos.UsageSummaryOptions{Days: 7})
	if err != nil {
		log.Printf("GetUsageSummary failed: %v", err)
	} else {
		fmt.Printf("\nUsage (7 days): %d total, %d today\n", summary.TotalRequests, summary.TotalToday)
	}
}
