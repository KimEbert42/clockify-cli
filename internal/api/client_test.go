package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClientDo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Api-Key") != "test-key" {
			t.Errorf("missing X-Api-Key header")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"ws-123","name":"Test Workspace"}`))
	}))
	defer server.Close()

	client := NewClient("test-key", server.URL, "ws-123")
	var result map[string]any
	err := client.Do(context.Background(), "GET", "/workspaces/ws-123", nil, &result)
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	if result["id"] != "ws-123" {
		t.Errorf("result[id] = %v, want ws-123", result["id"])
	}
}

func TestClientRetry(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := NewClient("test-key", server.URL, "ws-123",
		WithRetry(RetryPolicy{MaxRetries: 5, BaseDelay: 10 * time.Millisecond, MaxDelay: 100 * time.Millisecond,
			RetryableFn: func(statusCode int, err error) bool { return statusCode >= 500 },
		}),
	)
	var result map[string]any
	err := client.Do(context.Background(), "GET", "/test", nil, &result)
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	if attempts != 3 {
		t.Errorf("attempts = %d, want 3", attempts)
	}
}
