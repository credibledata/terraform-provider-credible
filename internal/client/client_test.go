package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	t.Run("with api key", func(t *testing.T) {
		c := NewClient("https://api.example.com", "ApiKey test-key", "my-org")
		if c.BaseURL != "https://api.example.com" {
			t.Errorf("expected BaseURL %q, got %q", "https://api.example.com", c.BaseURL)
		}
		if c.AuthHeader != "ApiKey test-key" {
			t.Errorf("expected AuthHeader %q, got %q", "ApiKey test-key", c.AuthHeader)
		}
		if c.Organization != "my-org" {
			t.Errorf("expected Organization %q, got %q", "my-org", c.Organization)
		}
		if c.HTTPClient == nil {
			t.Error("expected HTTPClient to be non-nil")
		}
	})

	t.Run("with bearer token", func(t *testing.T) {
		c := NewClient("https://api.example.com", "Bearer my-token", "")
		if c.AuthHeader != "Bearer my-token" {
			t.Errorf("expected AuthHeader %q, got %q", "Bearer my-token", c.AuthHeader)
		}
		if c.Organization != "" {
			t.Errorf("expected empty Organization, got %q", c.Organization)
		}
	})

	t.Run("trims trailing slash from base URL", func(t *testing.T) {
		c := NewClient("https://api.example.com/", "ApiKey k", "org")
		if c.BaseURL != "https://api.example.com" {
			t.Errorf("expected BaseURL without trailing slash, got %q", c.BaseURL)
		}
	})
}

func TestIsNotFound(t *testing.T) {
	t.Run("returns true for 404 error", func(t *testing.T) {
		err := fmt.Errorf("API error (HTTP 404): not found")
		if !IsNotFound(err) {
			t.Error("expected IsNotFound to return true for HTTP 404 error")
		}
	})

	t.Run("returns false for other errors", func(t *testing.T) {
		err := fmt.Errorf("API error (HTTP 500): internal server error")
		if IsNotFound(err) {
			t.Error("expected IsNotFound to return false for HTTP 500 error")
		}
	})

	t.Run("returns false for nil error", func(t *testing.T) {
		if IsNotFound(nil) {
			t.Error("expected IsNotFound to return false for nil error")
		}
	})
}

func TestDoRequest(t *testing.T) {
	t.Run("sets authorization header", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if got := r.Header.Get("Authorization"); got != "ApiKey test-key" {
				t.Errorf("expected Authorization header %q, got %q", "ApiKey test-key", got)
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{}`))
		}))
		defer server.Close()

		c := NewClient(server.URL, "ApiKey test-key", "org")
		_, _, err := c.doRequest("GET", "/test", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("constructs correct URL path", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			expected := "/api/v0/organizations/my-org"
			if r.URL.Path != expected {
				t.Errorf("expected path %q, got %q", expected, r.URL.Path)
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		c := NewClient(server.URL, "ApiKey k", "org")
		_, _, err := c.doRequest("GET", "/organizations/my-org", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("uses correct HTTP method", func(t *testing.T) {
		for _, method := range []string{"GET", "POST", "PATCH", "DELETE"} {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != method {
					t.Errorf("expected method %q, got %q", method, r.Method)
				}
				w.WriteHeader(http.StatusOK)
			}))

			c := NewClient(server.URL, "ApiKey k", "org")
			_, _, err := c.doRequest(method, "/test", nil)
			if err != nil {
				t.Fatalf("unexpected error for method %s: %v", method, err)
			}
			server.Close()
		}
	})

	t.Run("sets content-type for requests with body", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if got := r.Header.Get("Content-Type"); got != "application/json" {
				t.Errorf("expected Content-Type %q, got %q", "application/json", got)
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		c := NewClient(server.URL, "ApiKey k", "org")
		body := map[string]string{"name": "test"}
		_, _, err := c.doRequest("POST", "/test", body)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("does not set content-type for nil body", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if got := r.Header.Get("Content-Type"); got != "" {
				t.Errorf("expected no Content-Type header, got %q", got)
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		c := NewClient(server.URL, "ApiKey k", "org")
		_, _, err := c.doRequest("GET", "/test", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestDoJSON(t *testing.T) {
	t.Run("unmarshals successful response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"name": "test-org"})
		}))
		defer server.Close()

		c := NewClient(server.URL, "ApiKey k", "org")
		var result map[string]string
		err := c.doJSON("GET", "/test", nil, &result)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result["name"] != "test-org" {
			t.Errorf("expected name %q, got %q", "test-org", result["name"])
		}
	})

	t.Run("returns error for non-2xx status with API error message", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(APIError{Code: "invalid", Message: "bad request"})
		}))
		defer server.Close()

		c := NewClient(server.URL, "ApiKey k", "org")
		err := c.doJSON("GET", "/test", nil, nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if got := err.Error(); got != "API error (HTTP 400): bad request" {
			t.Errorf("unexpected error message: %s", got)
		}
	})

	t.Run("returns error for non-2xx status with raw body", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("internal error"))
		}))
		defer server.Close()

		c := NewClient(server.URL, "ApiKey k", "org")
		err := c.doJSON("GET", "/test", nil, nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if got := err.Error(); got != "API error (HTTP 500): internal error" {
			t.Errorf("unexpected error message: %s", got)
		}
	})

	t.Run("returns error for 404 that passes IsNotFound check", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(APIError{Code: "not_found", Message: "resource not found"})
		}))
		defer server.Close()

		c := NewClient(server.URL, "ApiKey k", "org")
		err := c.doJSON("GET", "/test", nil, nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !IsNotFound(err) {
			t.Error("expected IsNotFound to return true for 404 error")
		}
	})

	t.Run("handles nil result parameter", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		c := NewClient(server.URL, "ApiKey k", "org")
		err := c.doJSON("DELETE", "/test", nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
