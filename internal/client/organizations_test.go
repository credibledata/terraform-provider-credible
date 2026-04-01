package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateOrganization(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if r.URL.Path != "/api/v0/organizations" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}

			var body Organization
			json.NewDecoder(r.Body).Decode(&body)
			if body.Name != "test-org" {
				t.Errorf("expected name %q, got %q", "test-org", body.Name)
			}

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(Organization{
				Name:        "test-org",
				DisplayName: "Test Organization",
				CreatedAt:   "2025-01-01T00:00:00Z",
				UpdatedAt:   "2025-01-01T00:00:00Z",
			})
		}))
		defer server.Close()

		c := NewClient(server.URL, "ApiKey k", "org")
		result, err := c.CreateOrganization(&Organization{Name: "test-org"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Name != "test-org" {
			t.Errorf("expected name %q, got %q", "test-org", result.Name)
		}
		if result.DisplayName != "Test Organization" {
			t.Errorf("expected display_name %q, got %q", "Test Organization", result.DisplayName)
		}
	})

	t.Run("error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(APIError{Message: "already exists"})
		}))
		defer server.Close()

		c := NewClient(server.URL, "ApiKey k", "org")
		_, err := c.CreateOrganization(&Organization{Name: "test-org"})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestGetOrganization(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("expected GET, got %s", r.Method)
			}
			if r.URL.Path != "/api/v0/organizations/my-org" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(Organization{
				Name:        "my-org",
				DisplayName: "My Org",
				CreatedAt:   "2025-01-01T00:00:00Z",
				UpdatedAt:   "2025-01-01T00:00:00Z",
			})
		}))
		defer server.Close()

		c := NewClient(server.URL, "ApiKey k", "org")
		result, err := c.GetOrganization("my-org")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Name != "my-org" {
			t.Errorf("expected name %q, got %q", "my-org", result.Name)
		}
	})

	t.Run("not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(APIError{Message: "not found"})
		}))
		defer server.Close()

		c := NewClient(server.URL, "ApiKey k", "org")
		_, err := c.GetOrganization("missing")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !IsNotFound(err) {
			t.Error("expected IsNotFound to return true")
		}
	})
}

func TestUpdateOrganization(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "PATCH" {
				t.Errorf("expected PATCH, got %s", r.Method)
			}
			if r.URL.Path != "/api/v0/organizations/my-org" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(Organization{
				Name:        "my-org",
				DisplayName: "Updated Name",
				CreatedAt:   "2025-01-01T00:00:00Z",
				UpdatedAt:   "2025-01-02T00:00:00Z",
			})
		}))
		defer server.Close()

		c := NewClient(server.URL, "ApiKey k", "org")
		result, err := c.UpdateOrganization("my-org", &Organization{DisplayName: "Updated Name"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.DisplayName != "Updated Name" {
			t.Errorf("expected display_name %q, got %q", "Updated Name", result.DisplayName)
		}
	})
}
