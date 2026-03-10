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
	apiKey := os.Getenv("BEEOS_API_KEY")
	if apiKey == "" {
		log.Fatal("BEEOS_API_KEY is required")
	}

	client := beeos.NewClient(baseURL, apiKey)
	ctx := context.Background()

	// Search
	resp, err := client.Search(ctx, &beeos.SearchRequest{
		Query: "BeeOS AI agent platform",
		Limit: 5,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d results (took %dms):\n", resp.Meta.ResultCount, resp.Meta.DurationMs)
	for i, r := range resp.Results {
		fmt.Printf("  %d. %s\n     %s\n", i+1, r.Title, r.URL)
	}

	// Fetch
	if len(resp.Results) > 0 {
		fetchResp, err := client.Fetch(ctx, &beeos.FetchRequest{URL: resp.Results[0].URL})
		if err != nil {
			log.Printf("Fetch failed: %v", err)
			return
		}
		content := fetchResp.Content
		if len(content) > 200 {
			content = content[:200] + "..."
		}
		fmt.Printf("\nFetched content from %s:\n%s\n", fetchResp.SourceURL, content)
	}
}
