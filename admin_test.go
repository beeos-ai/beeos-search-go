package beeos_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	beeos "github.com/beeos-ai/beeos-search-go"
)

func TestAdminClient_CreateKey(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/admin/keys" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer mgmt-key" {
			t.Errorf("wrong auth header: %s", r.Header.Get("Authorization"))
		}

		var req beeos.CreateKeyRequest
		json.NewDecoder(r.Body).Decode(&req)

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(beeos.CreateKeyResponse{
			APIKey: beeos.APIKey{
				ID:        "key-123",
				KeyPrefix: "bs_abc",
				Name:      req.Name,
				RateLimit: req.RateLimit,
				CreatedAt: time.Now(),
			},
			Key: "bs_abcdef1234567890",
		})
	}))
	defer srv.Close()

	admin := beeos.NewAdminClient(srv.URL, "mgmt-key")
	resp, err := admin.CreateKey(context.Background(), &beeos.CreateKeyRequest{
		Name:      "test-key",
		RateLimit: 100,
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Key == "" {
		t.Error("expected plaintext key")
	}
	if resp.Name != "test-key" {
		t.Errorf("unexpected name: %s", resp.Name)
	}
}

func TestAdminClient_ListKeys(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"keys": []beeos.APIKey{
				{ID: "k1", Name: "key-1", KeyPrefix: "bs_aaa", RateLimit: 50},
				{ID: "k2", Name: "key-2", KeyPrefix: "bs_bbb", RateLimit: 100},
			},
		})
	}))
	defer srv.Close()

	admin := beeos.NewAdminClient(srv.URL, "mgmt-key")
	keys, err := admin.ListKeys(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
}

func TestAdminClient_RevokeKey(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/admin/keys/key-123" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	admin := beeos.NewAdminClient(srv.URL, "mgmt-key")
	if err := admin.RevokeKey(context.Background(), "key-123"); err != nil {
		t.Fatal(err)
	}
}

func TestAdminClient_GetUsageSummary(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/admin/usage/summary" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(beeos.UsageSummary{
			TotalRequests: 1234,
			TotalToday:    56,
			TotalWeek:     789,
			TotalMonth:    1234,
		})
	}))
	defer srv.Close()

	admin := beeos.NewAdminClient(srv.URL, "mgmt-key")
	summary, err := admin.GetUsageSummary(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if summary.TotalRequests != 1234 {
		t.Errorf("unexpected total: %d", summary.TotalRequests)
	}
}

func TestAdminClient_GetRecentLogs(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"logs": []beeos.UsageLog{
				{ID: 1, Endpoint: "/v1/search", Provider: "brave", StatusCode: 200},
			},
		})
	}))
	defer srv.Close()

	admin := beeos.NewAdminClient(srv.URL, "mgmt-key")
	logs, err := admin.GetRecentLogs(context.Background(), &beeos.RecentLogsOptions{Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(logs))
	}
}

func TestAdminClient_Login(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/admin/auth/login" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		auth := r.Header.Get("Authorization")
		if auth != "Bearer valid-token" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(beeos.APIError{Code: "unauthorized", Message: "invalid token"})
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"message": "ok"})
	}))
	defer srv.Close()

	admin := beeos.NewAdminClient(srv.URL, "valid-token")
	if err := admin.Login(context.Background()); err != nil {
		t.Fatal(err)
	}

	bad := beeos.NewAdminClient(srv.URL, "wrong-token")
	if err := bad.Login(context.Background()); err == nil {
		t.Fatal("expected error for wrong token")
	}
}

func TestAdminClient_NewClientFromAdmin(t *testing.T) {
	admin := beeos.NewAdminClient("https://search.beeos.ai", "mgmt-key")
	client := admin.NewClientFromAdmin("bs_my_new_key")
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}
