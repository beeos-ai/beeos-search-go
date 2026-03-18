package beeos_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	beeos "github.com/beeos-ai/beeos-search-go"
)

func TestClient_Search(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/search" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("missing or wrong auth header")
		}

		var req beeos.SearchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		if req.Query != "hello" {
			t.Errorf("unexpected query: %s", req.Query)
		}

		json.NewEncoder(w).Encode(beeos.SearchResponse{
			Results: []beeos.SearchResult{
				{Title: "Hello World", URL: "https://example.com", Snippet: "A greeting"},
			},
			Meta: beeos.SearchMeta{Query: "hello", ResultCount: 1, DurationMs: 42},
		})
	}))
	defer srv.Close()

	c := beeos.NewClient(srv.URL, "test-key")
	resp, err := c.Search(context.Background(), &beeos.SearchRequest{Query: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(resp.Results))
	}
	if resp.Results[0].Title != "Hello World" {
		t.Errorf("unexpected title: %s", resp.Results[0].Title)
	}
	if resp.Meta.DurationMs != 42 {
		t.Errorf("unexpected duration: %d", resp.Meta.DurationMs)
	}
}

func TestClient_Fetch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(beeos.FetchResponse{
			Content:     "<html>hi</html>",
			ContentType: "text/html",
			SourceURL:   "https://example.com",
		})
	}))
	defer srv.Close()

	c := beeos.NewClient(srv.URL, "test-key")
	resp, err := c.Fetch(context.Background(), &beeos.FetchRequest{URL: "https://example.com"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Content != "<html>hi</html>" {
		t.Errorf("unexpected content: %s", resp.Content)
	}
}

func TestClient_Providers(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"providers": []beeos.ProviderInfo{
				{ID: "brave", Name: "Brave", Regions: []string{"global"}, Status: "ok"},
			},
		})
	}))
	defer srv.Close()

	c := beeos.NewClient(srv.URL, "test-key")
	providers, err := c.Providers(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(providers) != 1 || providers[0].ID != "brave" {
		t.Errorf("unexpected providers: %+v", providers)
	}
}

func TestClient_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(beeos.APIError{Code: "unauthorized", Message: "invalid api key"})
	}))
	defer srv.Close()

	c := beeos.NewClient(srv.URL, "bad-key")
	_, err := c.Search(context.Background(), &beeos.SearchRequest{Query: "test"})
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr, ok := err.(*beeos.APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != 401 {
		t.Errorf("expected 401, got %d", apiErr.StatusCode)
	}
}

func TestClient_APIError_NestedBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":{"code":"UNAUTHORIZED","message":"invalid API key"}}`))
	}))
	defer srv.Close()

	c := beeos.NewClient(srv.URL, "bad-key")
	_, err := c.Search(context.Background(), &beeos.SearchRequest{Query: "test"})
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr, ok := err.(*beeos.APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != 401 {
		t.Errorf("expected status 401, got %d", apiErr.StatusCode)
	}
	if apiErr.Code != "UNAUTHORIZED" {
		t.Errorf("expected code UNAUTHORIZED, got %q", apiErr.Code)
	}
	if apiErr.Message != "invalid API key" {
		t.Errorf("expected message 'invalid API key', got %q", apiErr.Message)
	}
}

func TestClient_APIError_RateLimited(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error":{"code":"RATE_LIMITED","message":"rate limit exceeded"}}`))
	}))
	defer srv.Close()

	c := beeos.NewClient(srv.URL, "test-key")
	_, err := c.Search(context.Background(), &beeos.SearchRequest{Query: "test"})
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr, ok := err.(*beeos.APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != 429 {
		t.Errorf("expected status 429, got %d", apiErr.StatusCode)
	}
	if apiErr.Code != "RATE_LIMITED" {
		t.Errorf("expected code RATE_LIMITED, got %q", apiErr.Code)
	}
}

func TestClient_Healthz(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/healthz" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer srv.Close()

	c := beeos.NewClient(srv.URL, "")
	if err := c.Healthz(context.Background()); err != nil {
		t.Fatal(err)
	}
}
