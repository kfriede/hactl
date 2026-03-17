package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Error("expected Bearer token header")
		}
		if r.Header.Get("User-Agent") != userAgent {
			t.Errorf("expected User-Agent=%s, got %s", userAgent, r.Header.Get("User-Agent"))
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{
				{"id": "item-1", "name": "Test Item"},
			},
		})
	}))
	defer server.Close()

	c := NewClient(ClientConfig{
		Token:     "test-token",
		BaseURL:   server.URL,
		ErrWriter: discardWriter{},
	})

	var result map[string]any
	if err := c.GetJSON("", &result); err != nil {
		t.Fatalf("GetJSON failed: %v", err)
	}

	items, ok := result["items"].([]any)
	if !ok {
		t.Fatalf("expected items array, got %T", result["items"])
	}
	if len(items) != 1 {
		t.Errorf("expected 1 item, got %d", len(items))
	}
}

func TestClientRetry(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	c := NewClient(ClientConfig{
		Token:     "test-token",
		BaseURL:   server.URL,
		ErrWriter: discardWriter{},
	})

	var result map[string]string
	if err := c.GetJSON("", &result); err != nil {
		t.Fatalf("GetJSON failed after retries: %v", err)
	}

	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestClientAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code":    "not_found",
			"message": "Resource not found",
		})
	}))
	defer server.Close()

	c := NewClient(ClientConfig{
		BaseURL:   server.URL,
		ErrWriter: discardWriter{},
	})

	_, err := c.Get("")
	if err == nil {
		t.Fatal("expected error for 404 response")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.Code != "not_found" {
		t.Errorf("expected code 'not_found', got %q", apiErr.Code)
	}
}

type discardWriter struct{}

func (discardWriter) Write(p []byte) (int, error) { return len(p), nil }
